package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/", upload)
	http.ListenAndServe(":80", nil)
}

func upload(w http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	b, _ := io.ReadAll(request.Body)
	fmt.Fprintf(w, string(b))
}

// func HandleRequest(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {

// 	if !request.IsBase64Encoded || len(request.Body) == 0 {
// 		return events.LambdaFunctionURLResponse{Body: "Expecting a file", StatusCode: 418}, nil
// 	}

// 	filename, ok := request.Headers["filename"]
// 	if !ok {
// 		return events.LambdaFunctionURLResponse{Body: "Expecting a \"filename\" header", StatusCode: 400}, nil
// 	}

// 	file, err := base64.StdEncoding.DecodeString(request.Body)
// 	if err != nil {
// 		log.Println(err)
// 		return events.LambdaFunctionURLResponse{Body: "Failed to decode", StatusCode: 500}, nil
// 	}

// 	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
// 		Bucket:        aws.String(bucket_name),
// 		Key:           aws.String("input/" + filename),
// 		Body:          bytes.NewReader(file),
// 		ContentLength: int64(len(file)),
// 	})

// 	if err != nil {
// 		log.Println("Couldn't upload file: " + err.Error())
// 		return events.LambdaFunctionURLResponse{Body: "Failed to upload to s3", StatusCode: 500}, nil
// 	}

// 	return events.LambdaFunctionURLResponse{Body: "OK", StatusCode: 200}, nil
// }
