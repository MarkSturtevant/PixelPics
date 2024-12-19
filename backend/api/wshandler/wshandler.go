package wshandler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/fxamacker/cbor/v2"
)

func HandleWS(w http.ResponseWriter, r *http.Request) error {
	clientID := r.URL.Query().Get("id")

	ctx, cancel := context.WithCancelCause(r.Context())

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return fmt.Errorf("Error during websocket.Accept() %w\n", err)
	}

	var innerErr error
	defer func() {
		if tempErr := conn.CloseNow(); tempErr != nil {
			innerErr = fmt.Errorf("Error during conn.CloseNow() %w\n", tempErr)
		}
	}()

	defer cancel(errors.New("cancel websocket closed"))

	errCh := make(chan error, 1)

	textResChan := make(chan string)
	binResChan := make(chan []byte)

	go func() {
		defer cancel(errors.New("cancel websocket read closed"))

		for {
			msgType, data, err := conn.Read(ctx)
			if err != nil {
				errCh <- fmt.Errorf("Error during conn.Read() %w\n", err)
				return
			}

			if msgType == websocket.MessageBinary {
				var requestMessage WSRequestMessage
				err := cbor.Unmarshal(data, &requestMessage)
				if err != nil {
					errCh <- fmt.Errorf("Error during cbor.Unmarshal() %w\n", err)
					return
				}

				switch requestMessage.Type {
				case wsRequestMessageAwareness:
					if msg, ok := requestMessage.Message.(map[any]any); ok {
						newMsg := make(map[string]any)
						for k, v := range msg {
							if str, ok := k.(string); ok {
								newMsg[str] = v
							}
						}
						updateAwareness(clientID, newMsg)
					} else {
						errCh <- fmt.Errorf("Error during conn.Read() invalid message datatype: %T\n", requestMessage.Message)
					}
				case wsRequestMessageInvalid:
					fallthrough
				default:
					errCh <- fmt.Errorf("Error during conn.Read() invalid message type: %v\n", requestMessage.Type)
				}
			} else if msgType == websocket.MessageText {
				slog.Info(fmt.Sprintf("Received text message: %v\n", data))
			} else {
				errCh <- fmt.Errorf("Error during conn.Read() invalid message type: %v\n", msgType.String())
				return
			}
		}
	}()

	subscribeAwareness(ctx, clientID, binResChan)

	for {
		select {
		case textRes := <-textResChan:
			err := conn.Write(ctx, websocket.MessageText, []byte(textRes))
			if err != nil {
				errCh <- fmt.Errorf("Error during conn.Write() text %w\n", err)
				continue
			}
		case binRes := <-binResChan:
			err := conn.Write(ctx, websocket.MessageBinary, binRes)
			if err != nil {
				errCh <- fmt.Errorf("Error during conn.Write() bin %w\n", err)
				continue
			}
		case err := <-errCh:
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				slog.Info("Closing websocket connection", "code", closeErr.Code.String(), "reason", closeErr.Reason)
				return innerErr
			} else {
				return fmt.Errorf("Error during websocket endpoint: %w\n", err)
			}

		}
	}
}
