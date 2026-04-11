package zju

import (
	"time"
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

func (c *CoursesLogin) AnswerNumberRollcall(rid string, numberCode string) error {
	url := "https://courses.zju.edu.cn/api/rollcall/" + rid + "/answer_number_rollcall"

	// request body
	payload := map[string]string{
		"deviceId":   uuid.NewString(),
		"numberCode": numberCode,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := c.Fetch(ctx, url, &RequestOptions {
		Method: "PUT",
		Headers: http.Header {
			"Content-Type": {"application/json"},
		},
		Body: bytes.NewBuffer(bodyBytes),
	})
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Response:", string(respBody))

	return nil
}
