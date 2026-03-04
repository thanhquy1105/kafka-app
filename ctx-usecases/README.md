# Kafka Produce-with-Context Use Cases

Dự án này minh họa 4 kịch bản thực tế khi sử dụng `context` trong việc gửi tin nhắn Kafka bằng thư viện `franz-go`.

## 📌 Các Use-Case Cụ Thể

### 1. 🛡️ Chế độ Đồng bộ (Sync - /sync)
- **Mục tiêu:** Đảm bảo dữ liệu đã vào Kafka an toàn trước khi phản hồi người dùng.
- **Cách hoạt động:** Dùng `ProduceSync` hoặc đợi callback. Kết nối HTTP của User được giữ mở (Loading).
- **Phù hợp cho:** Giao dịch tài chính, đơn hàng, dữ liệu quan trọng cần xác nhận ngay.

### 2. ⚡ Chế độ Bất đồng bộ (Async - /async)
- **Mục tiêu:** Tối ưu tốc độ phản hồi cho người dùng.
- **Cách hoạt động:** Gọi `Produce` với `context.Background()` và trả về kết quả ngay. Việc gửi diễn ra ngầm.
- **Phù hợp cho:** Log hành vi, Tracking, Analytics (Dữ liệu có thể mất một chút cũng không sao).

### 3. ⏱️ Kiểm soát Thời gian (Timeout - /timeout)
- **Mục tiêu:** Bảo vệ hệ thống không bị treo khi Kafka chậm.
- **Cách hoạt động:** Dùng `context.WithTimeout`. Nếu Kafka không xác nhận trong thời gian định sẵn, hệ thống tự động hủy và báo "Bận" cho người dùng.
- **Phù hợp cho:** Các hệ thống có SLA (Service Level Agreement) khắt khe về thời gian phản hồi.

### 4. 🛑 Hủy bỏ khi User rời đi (Cancellation - /cancel)
- **Mục tiêu:** Tiết kiệm tài nguyên khi kết quả không còn cần thiết.
- **Cách hoạt động:** Truyền `r.Context()` vào Kafka. Nếu người dùng tắt trình duyệt (đóng tab), Context sẽ bị hủy -> Kafka Client ngừng việc cố gắng gửi/retry tin nhắn đó.
- **Phù hợp cho:** Các tác vụ nặng, tốn tài nguyên mà kết quả chỉ phục vụ cho chính người dùng đang chờ đó.

---

## 🛠️ Cách chạy demo

1. **Đảm bảo Kafka đang chạy:** (Dùng docker-compose trong thư mục gốc)
2. **Chạy server demo:**
   ```bash
   cd ctx-usecases
   go run main.go
   ```
3. **Thử nghiệm bằng trình duyệt hoặc CURL:**
   - `curl http://localhost:8080/sync`
   - `curl http://localhost:8080/async`
   - `curl http://localhost:8080/timeout`
   - `curl http://localhost:8080/cancel`
