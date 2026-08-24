import { expect, test, type APIRequestContext } from "@playwright/test";

const base = process.env.BASE_URL || "http://frontend";
const api = process.env.API_URL || "http://backend:8080";

async function login(request: APIRequestContext) {
  const res = await request.post(api + "/api/v1/auth/login", {
    data: { username: "photographer", password: "colorpixel" },
  });
  expect(res.ok()).toBeTruthy();
  const body = await res.json();
  expect(body.ok).toBeTruthy();
  return body.data.token as string;
}

test("critical flow: login library compare report", async ({ page, request }) => {
  const token = await login(request);
  const list = await request.get(api + "/api/v1/assets?page=1&page_size=4", {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(list.ok()).toBeTruthy();
  const payload = await list.json();
  expect(payload.data.total).toBeGreaterThan(0);
  const first = payload.data.items[0];
  expect(first.extraction_mode === "stream" || first.extraction_mode === "deferred").toBeTruthy();

  await page.goto(base + "/login");
  await page.getByRole("heading", { name: "暗房观测台" }).waitFor();
  await page.locator('input[type="password"]').fill("colorpixel");
  await page.getByRole("button", { name: "进入工作台" }).click();
  await expect(page.getByRole("link", { name: "ColorPixel" })).toBeVisible();
  await expect(page.getByText("预览级 (Embedded JPEG)").first()).toBeVisible();

  await page.getByRole("link", { name: "比对墙" }).click();
  await expect(page.getByText("预览级 (Embedded JPEG)").first()).toBeVisible();
  await page.getByRole("button", { name: "四镜" }).click();

  await page.getByRole("link", { name: "直方图" }).click();
  await expect(page.getByRole("heading", { name: "RGB 通道大屏" })).toBeVisible();

  await page.getByRole("link", { name: "挂机镜" }).click();
  await expect(page.getByRole("heading", { name: "黄金挂机镜" })).toBeVisible();
});
