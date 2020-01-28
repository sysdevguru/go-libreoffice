FROM lambci/lambda:build-go1.x AS build_base_golang

ENV GO111MODULE="on"
WORKDIR /work/go
COPY . .
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o go-libreoffice
RUN zip -9yr lambda.zip .

CMD aws lambda update-function-code --function-name pdf_to_docx_function --zip-file fileb://lambda.zip