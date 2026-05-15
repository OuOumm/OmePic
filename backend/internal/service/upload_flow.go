package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"omepic/backend/internal/model"
	"omepic/backend/internal/storage"
)

type uploadFlow struct {
	service *ImageService
	ctx     context.Context
	input   UploadInput
	clock   func() time.Time
}

type uploadRuntimePolicy struct {
	settings     RuntimeSettings
	maxBytes     int64
	avifSettings AVIFConversionSettings
}

// uploadTransaction owns the upload-time resource bundle and commit state machine:
// runtime policy snapshot, prepared original source, resolved storage instance,
// scoped original-byte MD5 key, and hash-lock lifetime. uploadFlow.Run should only
// create a transaction and ask it to commit; duplicate reuse, new physical writes,
// cleanup, and Redis/SQLite consistency stay implementation details here.
type uploadTransaction struct {
	flow     uploadFlow
	policy   uploadRuntimePolicy
	source   preparedUploadSource
	storage  storage.ResolvedProvider
	md5Key   model.MD5MappingKey
	unlockFn func()
}

type physicalUploadResult struct {
	uid        string
	storedPath string
	size       int64
}

func (s *ImageService) newUploadFlow(ctx context.Context, input UploadInput) uploadFlow {
	return uploadFlow{
		service: s,
		ctx:     ctx,
		input:   input,
		clock: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (f uploadFlow) Run() (UploadOutput, error) {
	tx, err := f.beginTransaction()
	if err != nil {
		return UploadOutput{}, err
	}
	defer tx.cleanup()

	return tx.Commit()
}

func (f uploadFlow) beginTransaction() (*uploadTransaction, error) {
	if strings.TrimSpace(f.input.Token) == "" {
		return nil, ErrMissingToken
	}
	if err := f.service.ensureIPAllowed(f.ctx, f.input.IPAddress); err != nil {
		return nil, err
	}

	policy := f.runtimePolicy()
	if policy.settings.MaintenanceMode {
		return nil, WithUserMessage(ErrInvalidInput, policy.settings.EffectiveMaintenanceMessage())
	}

	source, err := f.service.prepareUploadSource(f.input, policy.maxBytes)
	if err != nil {
		return nil, err
	}
	cleanupOnError := true
	defer func() {
		if cleanupOnError {
			source.Cleanup()
		}
	}()

	if source.size == 0 {
		return nil, emptyUploadError(policy.settings)
	}
	if !runtimeSettingsAllowsMIME(policy.settings, f.input.MIMEType) {
		return nil, WithUserMessage(ErrInvalidInput, "file MIME type is not allowed")
	}

	resolved, err := f.service.resolveUploadStorage(f.input.StorageKey, policy.settings.AllowStorageSelect)
	if err != nil {
		return nil, err
	}

	md5Hash := strings.TrimSpace(strings.ToLower(source.originalMD5))
	if md5Hash == "" {
		return nil, fmt.Errorf("%w: original md5 is required after upload source preparation", ErrDependencyUnavailable)
	}

	md5Key := model.NewMD5MappingKey(resolved.Config.StorageKey, md5Hash)
	unlockHash := f.service.hashLocks.Lock(md5Key.MutexScope())
	cleanupOnError = false
	return &uploadTransaction{
		flow:     f,
		policy:   policy,
		source:   source,
		storage:  resolved,
		md5Key:   md5Key,
		unlockFn: unlockHash,
	}, nil
}

func (f uploadFlow) runtimePolicy() uploadRuntimePolicy {
	settings := f.service.currentRuntimeSettings()
	return uploadRuntimePolicy{
		settings:     settings,
		maxBytes:     settings.MaxUploadSizeBytes(),
		avifSettings: avifConversionSettingsFromRuntime(settings),
	}
}

func (tx *uploadTransaction) cleanup() {
	if tx.unlockFn != nil {
		tx.unlockFn()
	}
	tx.source.Cleanup()
}

func (tx *uploadTransaction) Commit() (UploadOutput, error) {
	existing, err := tx.findReusableObject()
	if err != nil {
		return UploadOutput{}, err
	}

	uid, err := tx.flow.service.generateUID()
	if err != nil {
		return UploadOutput{}, fmt.Errorf("%w: uid generation failed", ErrDependencyUnavailable)
	}

	if existing != nil {
		return tx.commitDuplicate(uid, *existing)
	}
	return tx.commitNewObject(uid)
}

func (tx *uploadTransaction) findReusableObject() (*model.ImageRecord, error) {
	return tx.flow.service.md5Mappings().FindReusableObject(tx.flow.ctx, tx.md5Key)
}

func (tx *uploadTransaction) commitDuplicate(uid string, existing model.ImageRecord) (UploadOutput, error) {
	record := tx.duplicateRecord(uid, existing)
	if err := tx.commitImageRecord(record); err != nil {
		return UploadOutput{}, err
	}
	return buildUploadOutput(record, tx.flow.input.BaseURL, tx.flow.input.OriginalFilename, true), nil
}

func (tx *uploadTransaction) duplicateRecord(uid string, existing model.ImageRecord) model.ImageRecord {
	return model.ImageRecord{
		UID:            uid,
		Token:          tx.flow.input.Token,
		StorageKey:     existing.StorageKey,
		StorageBackend: existing.StorageBackend,
		FilePath:       existing.FilePath,
		MIMEType:       existing.MIMEType,
		Size:           existing.Size,
		MD5Hash:        existing.MD5Hash,
		IPAddress:      tx.flow.input.IPAddress,
		CreatedAt:      tx.flow.clock(),
	}
}

func (tx *uploadTransaction) commitNewObject(uid string) (UploadOutput, error) {
	physical, err := tx.writePhysicalObject(uid)
	if err != nil {
		return UploadOutput{}, err
	}
	return tx.commitNewPhysical(physical)
}

func (tx *uploadTransaction) writePhysicalObject(uid string) (physicalUploadResult, error) {
	sourceReader, err := tx.source.Open()
	if err != nil {
		return physicalUploadResult{}, fmt.Errorf("%w: failed to open upload source", ErrDependencyUnavailable)
	}
	defer sourceReader.Close()

	objectKey := storage.BuildObjectKey(uid, publicImageExtension)
	convertedSize, storedPath, err := tx.flow.service.saveConvertedAVIF(tx.flow.ctx, tx.storage.Provider, objectKey, sourceReader, tx.policy.avifSettings)
	if err != nil {
		return physicalUploadResult{}, err
	}
	return physicalUploadResult{uid: uid, storedPath: storedPath, size: convertedSize}, nil
}

func (tx *uploadTransaction) commitNewPhysical(physical physicalUploadResult) (UploadOutput, error) {
	record := tx.newPhysicalRecord(physical)
	if err := tx.insertImageRecord(record); err != nil {
		_ = tx.storage.Provider.Delete(tx.flow.ctx, physical.storedPath)
		return UploadOutput{}, err
	}
	if err := tx.writeUIDCache(record); err != nil {
		return UploadOutput{}, err
	}
	if err := tx.flow.service.md5Mappings().RememberNewPhysical(tx.flow.ctx, tx.md5Key, physical.uid); err != nil {
		return UploadOutput{}, err
	}
	return buildUploadOutput(record, tx.flow.input.BaseURL, tx.flow.input.OriginalFilename, false), nil
}

func (tx *uploadTransaction) newPhysicalRecord(physical physicalUploadResult) model.ImageRecord {
	return model.ImageRecord{
		UID:            physical.uid,
		Token:          tx.flow.input.Token,
		StorageKey:     tx.storage.Config.StorageKey,
		StorageBackend: tx.storage.Config.Backend,
		FilePath:       physical.storedPath,
		MIMEType:       publicImageMIMEType,
		Size:           physical.size,
		MD5Hash:        tx.md5Key.MD5Hash,
		IPAddress:      tx.flow.input.IPAddress,
		CreatedAt:      tx.flow.clock(),
	}
}

func (tx *uploadTransaction) commitImageRecord(record model.ImageRecord) error {
	if err := tx.insertImageRecord(record); err != nil {
		return err
	}
	return tx.writeUIDCache(record)
}

func (tx *uploadTransaction) insertImageRecord(record model.ImageRecord) error {
	if err := tx.flow.service.repo.InsertImage(tx.flow.ctx, record); err != nil {
		return mapRepoError(err)
	}
	return nil
}

func (tx *uploadTransaction) writeUIDCache(record model.ImageRecord) error {
	if err := tx.flow.service.imageCache.SetImage(tx.flow.ctx, record); err != nil {
		return fmt.Errorf("%w: redis uid write failed", ErrDependencyUnavailable)
	}
	return nil
}

func emptyUploadError(settings RuntimeSettings) error {
	if settings.MaxUploadSizeBytes() > 0 {
		return WithUserMessage(ErrInvalidInput, fmt.Sprintf("file size must be between 1 byte and %d MB", settings.MaxUploadSizeMB))
	}
	return WithUserMessage(ErrInvalidInput, "file size must be greater than 0 bytes")
}

type avifStreamSaveResult struct {
	storedPath string
	err        error
}

type avifStreamConversion struct {
	encoder  func(io.Reader, io.Writer, AVIFConversionSettings) error
	provider storage.Provider
	settings AVIFConversionSettings
}

func (c avifStreamConversion) save(ctx context.Context, objectKey string, source io.Reader) (int64, string, error) {
	pipeReader, pipeWriter := io.Pipe()
	counting := &countingWriter{writer: pipeWriter}
	encodeErrCh := make(chan error, 1)
	saveResultCh := make(chan avifStreamSaveResult, 1)

	go func() {
		err := c.encoder(source, counting, c.settings)
		if err != nil {
			_ = pipeWriter.CloseWithError(err)
			encodeErrCh <- err
			return
		}
		encodeErrCh <- pipeWriter.Close()
	}()

	go func() {
		storedPath, err := c.provider.SaveStream(ctx, objectKey, pipeReader, -1, publicImageMIMEType)
		saveResultCh <- avifStreamSaveResult{storedPath: storedPath, err: err}
	}()

	saveResult := <-saveResultCh
	if saveResult.err != nil {
		_ = pipeReader.Close()
		_ = pipeWriter.CloseWithError(saveResult.err)
	}

	encodeErr := <-encodeErrCh
	_ = pipeReader.Close()

	if err := avifStreamError(encodeErr, saveResult.err); err != nil {
		return 0, "", err
	}
	return counting.size, saveResult.storedPath, nil
}

func avifStreamError(encodeErr error, saveErr error) error {
	if encodeErr != nil {
		if saveErr == nil || errors.Is(saveErr, encodeErr) || !errors.Is(encodeErr, io.ErrClosedPipe) {
			return encodeErr
		}
	}
	if saveErr != nil {
		return fmt.Errorf("%w: failed to persist file", ErrDependencyUnavailable)
	}
	return nil
}
