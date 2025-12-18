## Ứng dụng lấy mã 2FA

- Sử dụng GO
- Phục vụ đa luồng hiệu suất cao
- Cung cấp API để trả về mã 2FA cho người dùng

**Demo:** https://2fa.scapbot.net

## API Endpoints

### POST /api/totp (Khuyên dùng)

Tạo mã TOTP 2FA từ secret key. **Hỗ trợ secret có khoảng trắng mà không cần URL encode.**

**Request:**
```bash
curl -X POST "https://2fa.scapbot.net/api/totp" \
  -H "Content-Type: application/json" \
  -d '{"secret":"PL5LQ LSH7L R3BMJ FYTJ6 G6MSZ IGJTXZJ"}'
```

### GET /api/totp

**Request:**
```bash
curl "https://2fa.scapbot.net/api/totp?secret=PL5LQLSH7LR3BMJFYTJ6G6MSZIGJTXZJ"
```

**Parameters:**
| Tham số | Mô tả |
|---------|-------|
| `secret` | Secret key dạng Base32 (có hoặc không có khoảng trắng) |

**Response thành công:**
```json
{
  "success": true,
  "code": "416681",
  "remaining": 21
}
```

| Field | Mô tả |
|-------|-------|
| `success` | Trạng thái xử lý |
| `code` | Mã 2FA 6 chữ số |
| `remaining` | Số giây còn lại trước khi mã hết hạn |

**Response lỗi:**
```json
{
  "success": false,
  "error": "invalid secret key: illegal base32 data at input byte 7"
}
```

### GET /health

Kiểm tra trạng thái server.

```bash
curl "https://2fa.scapbot.net/health"
```

**Response:**
```json
{
  "status": "ok"
}
```

## Chạy với Docker

```bash
# Build image
docker build -t 2fa-api .

# Run container
docker run -d --name 2fa-api -p 8080:8080 2fa-api

# Stop container
docker rm -f 2fa-api
```

## Deploy lên VPS (aaPanel)

```bash
# 1. Upload code lên VPS

# 2. Chạy với docker-compose
docker-compose up -d --build

# 3. Cấu hình reverse proxy trên aaPanel
#    - Tạo website: 2fa.scapbot.net
#    - Reverse proxy đến: http://127.0.0.1:7842
#    - Bật SSL
```