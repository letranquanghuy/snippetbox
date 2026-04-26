package main

import (
	"bytes"
	"net/http"
	"runtime/debug"
	"time"

	"fmt"
)

// The serverError helper writes a log entry at Error level (including the request
// method and URI as attributes), then sends a generic 500 Internal Server Error
// response to the user.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		// Use debug.Stack() to get the stack trace. This returns a byte slice, which
		// we need to convert to a string so that it's readable in the log entry.
		//debug.Stack() từ package runtime/debug trả về stack trace của goroutine hiện tại dưới dạng []byte.
		// Khi có lỗi server, mình muốn biết luồng thực thi dẫn đến lỗi đó — hàm nào gọi hàm nào, ở dòng mấy.
		// 		goroutine 1 [running]:
		// 		runtime/debug.Stack()
		// 		    /usr/local/go/src/runtime/debug/stack.go:24 +0x5b
		// 		main.(*application).serverError(...)
		// 		    /home/user/snippetbox/cmd/web/helpers.go:12
		// 		main.(*application).home(...)
		// 		    /home/user/snippetbox/cmd/web/handlers.go:25
		// 		...
		//Nếu không có debug.Stack(), log chỉ ghi được:
		// 		ERROR method=GET uri=/foo stack=""
		// 		some error message
		trace = string(debug.Stack())
	)

	// Include the trace in the log entry
	app.logger.Error(err.Error(), "method", method, "uri", uri, "stack", trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	// Retrieve the appropriate template set from the cache based on the page name.
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// option 1: simply execute the template straight to http.ResponseWriter
	// WriteHeader gửi status code ngay
	// Nếu ExecuteTemplate lỗi giữa chừng thì không thể đổi lại status 500 nữa vì header đã gửi rồi
	// User có thể nhận được response nửa vời — một phần HTML + lỗi 500 lẫn lộn
	// w.WriteHeader(status)

	// // Execute the template, passing in any dynamic data.
	// err := ts.ExecuteTemplate(w, "base", data)
	// if err != nil {
	// 	app.serverError(w, r, err)
	// }

	// option 2: safe way to execute template vào buffer trước, rồi mới gửi cho client
	// Render template vào buffer trước, chưa gửi gì cho client
	// Nếu lỗi → gửi 500 sạch sẽ
	// Nếu thành công → mới WriteHeader rồi buf.WriteTo(w) gửi toàn bộ HTML hoàn chỉnh
	// Initialize a new buffer.
	buf := new(bytes.Buffer)
	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our serverError() helper
	// and then return.
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	// If the template is written to the buffer without any errors, we are safe
	// to go ahead and write the HTTP status code to http.ResponseWriter.
	w.WriteHeader(status)
	// Write the contents of the buffer to the http.ResponseWriter. Note: this
	// is another time where we pass our http.ResponseWriter to a function that
	// takes an io.Writer.
	buf.WriteTo(w)
}

// Create an newTemplateData() helper, which returns a templateData struct
// initialized with the current year. Note that we're not using the *http.Request
// parameter here at the moment, but we will do later in the book.
func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear: time.Now().Year(),
	}
}
