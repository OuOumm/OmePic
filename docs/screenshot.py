"""Take screenshots of OmePic for README."""
import asyncio
from playwright.async_api import async_playwright

BASE_URL = "http://localhost:8080"
OUTPUT_DIR = "docs/screenshots"

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch()
        context = await browser.new_context(
            viewport={"width": 1440, "height": 900},
            locale="zh-CN",
        )
        page = await context.new_page()

        # 1. Homepage - Upload page
        print("[1/4] Capturing homepage...")
        await page.goto(BASE_URL, wait_until="networkidle")
        await page.wait_for_timeout(1000)
        await page.screenshot(path=f"{OUTPUT_DIR}/home.png", full_page=False)

        # 2. Admin login page
        print("[2/4] Capturing admin login...")
        await page.goto(f"{BASE_URL}/admin", wait_until="networkidle")
        await page.wait_for_timeout(1000)
        await page.screenshot(path=f"{OUTPUT_DIR}/admin-login.png", full_page=False)

        # 3. Login and go to admin dashboard
        print("[3/4] Logging in to admin...")
        # Try to find and fill the password input
        pwd_input = page.locator('input[type="password"]')
        if await pwd_input.count() > 0:
            await pwd_input.fill("admin123")
            # Find submit button
            submit_btn = page.locator('button[type="submit"]')
            if await submit_btn.count() > 0:
                await submit_btn.click()
            else:
                await pwd_input.press("Enter")
            await page.wait_for_timeout(2000)

        # 4. Admin dashboard / images page
        print("[4/4] Capturing admin dashboard...")
        await page.goto(f"{BASE_URL}/admin", wait_until="networkidle")
        await page.wait_for_timeout(1000)
        await page.screenshot(path=f"{OUTPUT_DIR}/admin-dashboard.png", full_page=False)

        await browser.close()
        print(f"\nDone! Screenshots saved to {OUTPUT_DIR}/")
        print("  - home.png")
        print("  - admin-login.png")
        print("  - admin-dashboard.png")

asyncio.run(main())
