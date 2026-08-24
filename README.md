# ColorPixel · Mini 色彩像素墙

多相机数字资产色彩管理、高保真 RAW 画幅对比与镜头 EXIF 元数据审计中心。

## 1. 如何启动

```bash
docker compose up --build -d
```

浏览器访问 http://localhost:28381 。首次启动会写入合成 RAW 样本并异步切瓦片，后端健康检查 `start_period` 为 120 秒。

## 2. 使用说明

使用测试账号登录后进入资产墙。可按机身、镜头、文件名筛选，勾选 2–4 张进入比对墙，滚轮缩放、拖拽平移默认完全同步。直方图页跟随当前选中资产刷新 RGB 通道。挂机镜页给出过去 365 天（`DateTimeOriginal`，GMT+8）的加权素质报告与推导依据。界面常驻「预览级 (Embedded JPEG)」徽标，比对源为 RAW 内嵌全分辨率 JPEG 预览，不是传感器 Bayer 数据。

## 3. 服务列表及API说明

| 服务 | 地址 |
|---|---|
| 工作台 | http://localhost:28381 |
| API（经 nginx 同源反代） | http://localhost:28381/api/v1 |
| 健康检查 | http://localhost:28381/api/v1/health |
| PostgreSQL | localhost:28383（仅本机调试） |

完整契约、请求/响应示例与错误码见 `docs/API.md`。

## 4. 测试账号

- 用户名：`photographer`
- 密码：`colorpixel`

## 5. 题目内容

使用 Go 实现职业修图团队的 RAW 资产管理中心：纯 Go 手写字节流扫描器抽取 EXIF 与内嵌 JPEG，网页端双镜/四镜同步像素比对、实时 RGB 直方图，以及基于一年 EXIF 的黄金挂机镜加权评价。禁止第三方 C 库。规模约 36+ Go 文件、7k–9k 行。

## 6. 项目结构

```
backend/           Go 服务、解析器、瓦片、镜头算法
frontend-user/    Vue 3 工作台
tests/            API smoke + Playwright
docs/             需求、路线图、API、设计、QA、审计
docker-compose.yml
```

## 7. API 模拟与切换指南

本项目**没有**第三方计费 API。唯一 Mock 是仓库内的 RAW **样本数据**，不是解析逻辑。

- **真实通路**：`internal/raw` 按 TIFF/EP 与 ISO-BMFF 规范解析。将真实 `.CR3/.NEF/.ARW/.CR2/.DNG` 放到上传接口或 `DATA_DIR/raw` 即可走同一解析器。
- **Mock 数据**：`SAMPLE_MODE=1`（默认）在空库时由 `internal/sample` 合成最小合法容器（真实 IFD/Box、真实 tag、真实 JPEG）。`SAMPLE_MODE=0` 关闭自动播种。
- **延迟抽取样例**：目录中的 `DEFERRED_PREVIEW.NEF` 将预览放在 16MB 窗口之外，API 必须返回 `extraction_mode=deferred`。
- 切换：环境变量 `SAMPLE_MODE`；投放真实相机文件无需改代码。
