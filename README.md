# Golang LibreOffice wrapper
Simple AWS Lambda function that converts uploaded pdf files in S3 bucket to docx and returns url of the converted file.  
It uses soffice binary for converting in `--headless` mode.  

## Prerequisites
| Environment variable  | Description                                |
|:----------------------|:-------------------------------------------|
| AWS_REGION            | AWS Region                                 |
| AWS_DEFAULT_REGION    | AWS Default Region                         |
| AWS_ACCESS_KEY_ID     | AWS Access Key Id                          |
| AWS_SECRET_ACCESS_KEY | AWS Secret Access Key                      |

AWS account  
AWS lambda function created in AWS console  

## Testing locally with lambci/docker-lambda
### Run lambda function
```sh
go build -o go-libreoffice main.go
docker run --rm -e AWS_REGION -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e DOCKER_LAMBDA_STAY_OPEN=1 -p 9001:9001 -v "$PWD":/var/task:ro,delegated lambci/lambda:go1.x go-libreoffice
```
### Create S3 bucket put event
```sh
curl --header "Content-Type: application/json" --request POST --data @body.json http://localhost:9001/2015-03-31/functions/go-libreoffice/invocations
```
body.json is the S3 bucket put event information.  
Here is a body.json you can use for testing.  
Assuming that you already have created Lambda function and uploaded a PDF file called test.pdf for this testing  
```sh
{
  "Records": [
    {
      "eventVersion": "2.0",
      "eventSource": "aws:s3",
      "awsRegion": "[YOUR_REGION]",
      "eventTime": "1970-01-01T00:00:00.000Z",
      "eventName": "ObjectCreated:Put",
      "userIdentity": {
        "principalId": "EXAMPLE"
      },
      "requestParameters": {
        "sourceIPAddress": "127.0.0.1"
      },
      "responseElements": {
        "x-amz-request-id": "EXAMPLE123456789",
        "x-amz-id-2": "EXAMPLE123/5678abcdefghijklambdaisawesome/mnopqrstuvwxyzABCDEFGH"
      },
      "s3": {
        "s3SchemaVersion": "1.0",
        "configurationId": "testConfigRule",
        "bucket": {
          "name": "[YOUR_LAMBDA_FUNCTION_NAME]",
          "ownerIdentity": {
            "principalId": "EXAMPLE"
          },
          "arn": "arn:aws:s3:::example-bucket"
        },
        "object": {
          "key": "test.pdf",
          "size": 1024,
          "eTag": "0123456789abcdef0123456789abcdef",
          "sequencer": "0A1B2C3D4E5F678901"
        }
      }
    }
  ]
}
```

## Deploying with Dockerfile
```sh
docker build -t lambdaFunc .
docker run --rm -e AWS_DEFAULT_REGION -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY lambdaFunc
```
