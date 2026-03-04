package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	brokers = []string{"localhost:9092"}
	topic   = "use-case-topic"
	client  *kgo.Client

	// Dùng để giả lập việc nhận phản hồi từ Consumer khác
	// Trong thực tế, bạn sẽ có một Consumer thực thụ nghe topic "replies"
	pendingReplies = make(map[string]chan string)
	mu             sync.Mutex
)

func initKafka() {
	var err error
	client, err = kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.DefaultProduceTopic(topic),
	)
	if err != nil {
		log.Fatalf("Unable to create kafka client: %v", err)
	}
}

func main() {
	initKafka()
	defer client.Close()

	// Giả lập một Consumer chạy ngầm để xử lý và "trả lời"
	go simulateWorker()

	r := gin.Default()

	r.GET("/sync", handleSync)
	r.GET("/async", handleAsync)
	r.GET("/timeout", handleTimeout)
	r.GET("/cancel", handleCancel)

	// --- USE CASE MỚI: REQUEST-REPLY ---
	r.GET("/request-reply", handleRequestReply)

	fmt.Println("🚀 Gin Use-Case Server started at :8080")
	r.Run(":8080")
}

// Case 5: Request-Reply - Waiting for Consumer to finish
func handleRequestReply(c *gin.Context) {
	// 1. Tạo một Correlation ID duy nhất để biết phản hồi nào là của request nào
	requestID := uuid.New().String()

	// 2. Tạo một channel để chờ kết quả
	replyChan := make(chan string, 1)
	mu.Lock()
	pendingReplies[requestID] = replyChan
	mu.Unlock()

	// Đảm bảo dọn dẹp khi kết thúc
	defer func() {
		mu.Lock()
		delete(pendingReplies, requestID)
		mu.Unlock()
	}()

	// 3. Thiết lập timeout: User chỉ chờ tối đa 5 giây cho Worker xử lý
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 4. Gửi yêu cầu vào Kafka kèm theo ID
	record := &kgo.Record{
		Key:   []byte(requestID),
		Value: []byte("Process this heavy task, please!"),
	}
	client.Produce(ctx, record, nil)

	// 5. CHỜ KẾT QUẢ TỪ CONSUMER (hoặc chờ Timeout/Cancel)
	select {
	case result := <-replyChan:
		// Worker đã xử lý xong!
		c.JSON(http.StatusOK, gin.H{
			"status":     "completed",
			"request_id": requestID,
			"result":     result,
		})
	case <-ctx.Done():
		// Quá thời gian chờ hoặc User đã tắt trình duyệt
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error": "Worker is taking too long to process. Please check back later.",
			})
		} else {
			log.Println("User disconnected, stopping wait for reply.")
		}
	}
}

// Giả lập một Worker nhận tin nhắn từ Kafka và "phản hồi" sau 2 giây
func simulateWorker() {
	for {
		// Trong thực tế, chỗ này sẽ là client.PollFetches từ Kafka
		// Ở đây mình giả lập việc quét các request đang chờ mỗi giây
		time.Sleep(2 * time.Second)

		mu.Lock()
		for id, ch := range pendingReplies {
			// Giả lập xử lý xong và gửi vào channel "phiên dịch" kết quả
			select {
			case ch <- fmt.Sprintf("Finished processing at %s", time.Now().Format("15:04:05")):
				log.Printf("Worker: Handled request %s", id)
			default:
			}
		}
		mu.Unlock()
	}
}

// --- CÁC HANDLER CŨ (Giữ nguyên để bạn so sánh) ---

func handleSync(c *gin.Context) {
	// Gin Context implements context.Context interface
	ctx := c.Request.Context()

	record := &kgo.Record{Value: []byte("Important Financial Data via Gin")}

	results := client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Kafka confirmed receipt. Safe to proceed.",
		"offset":  record.Offset,
	})
}

// Case 2: Asynchronous - Fire and Forget
func handleAsync(c *gin.Context) {
	// IMPORTANT: Don't use c.Request.Context() because Gin will cancel it
	// as soon as this handler returns. Use a detached context for BG work.
	bgCtx := context.Background()

	record := &kgo.Record{Value: []byte("Analytics Data via Gin")}

	client.Produce(bgCtx, record, func(r *kgo.Record, err error) {
		if err != nil {
			log.Printf("BG Produce Failed: %v", err)
		} else {
			log.Printf("BG Produce Success: offset %d", r.Offset)
		}
	})

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"message": "Request accepted. Processing in background.",
	})
}

// Case 3: Timeout - Protective handling
func handleTimeout(c *gin.Context) {
	// Set a very short timeout for this specific operation
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Millisecond)
	defer cancel()

	record := &kgo.Record{Value: []byte("Ticking Bomb via Gin")}

	results := client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"reason": "Kafka too slow or system busy",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Case 4: Cancellation - User walks away
func handleCancel(c *gin.Context) {
		ctx := c.Request.Context()
	
	log.Println("Bắt đầu xử lý chuỗi 10 tin nhắn Kafka...")
	for i := 1; i <= 10; i++ {
		// KIỂM TRA: Nếu User đã ngắt kết nối thì dừng ngay, không gửi tiếp nữa
		select {
		case <-ctx.Done():
			log.Printf("🏮 User đã thoát ở bước %d. Hủy bỏ tác vụ!", i)
			return // Thoát handler, log sẽ không in thêm bước nào nữa
		default:
			// Giả lập việc gửi tin nhắn vào Kafka
			record := &kgo.Record{
				Value: []byte(fmt.Sprintf("Batch Record #%d", i)),
			}
			
			// Truyền ctx vào Produce để Kafka Client cũng biết mà dừng nếu cần
			client.Produce(ctx, record, func(_ *kgo.Record, err error) {
				if err != nil {
					log.Printf("Lỗi gửi tin %d: %v", i, err)
				}
			})
			// Giả lập xử lý nặng (ví dụ: truy vấn DB hoặc tính toán) mất 500ms mỗi bước
			time.Sleep(500 * time.Millisecond)
			log.Printf("✅ Đã gửi tin nhắn thứ %d", i)
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Đã gửi toàn bộ 10 tin nhắn thành công!"})
}

func record() *kgo.Record { return &kgo.Record{Value: []byte("data")} }
