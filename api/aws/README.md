# AWS Infrastructure

This project uses nested stacks. To deploy a project with nested stacks we must first upload the stack templates
to an S3 bucket. This bucket is tracked in the `deploy.yml` stack.

## Steps to deploy the Infrastructure

### 1. Deploy the `deploy` stack

The deploy bucket stack is a regular stack so to deploy just:

```
aws cloudformation deploy --profile cycleforlisbon --template-file ./deploy.yaml --stack-name cycleforlisbon-deploy --region us-east-1
```

It _has_ to be deployed in us-east-1. Otherwise cloudfront complains.

### 2. Add DNS Records

The stack will stay in deployment until the TLS certificates are emitted.

Go to Amazon ACM grab the CNAME records and insert them in the Domain Name DNS.

### 3. Deploy the main stack

To deploy the main infrastructure we can use the following command

```
aws cloudformation package --profile cycleforlisbon --template-file ./main.yaml --s3-bucket cycleforlisbon-dev-deploy --region us-east-1 --output-template-file tmp-template.yaml &&
   aws cloudformation deploy --profile cycleforlisbon --region eu-west-1 --capabilities CAPABILITY_NAMED_IAM --template-file tmp-template.yaml --stack-name cycleforlisbon &&
   rm tmp-template.yaml
```

### 4. Create Dex DB

Once RDS Is up, connect to it and create the dex database with

```
CREATE DATABASE dex;
```

### 5. Upload the docker containers to ECR

The deploy will stay stuck while a docker image isn't uploaded to ECR.

```
STACKNAME=cycleforlisbon ./aws/scripts/deploy-api.sh
```

For manual deployment steps see the following:

See this guide https://docs.aws.amazon.com/AmazonECR/latest/userguide/docker-push-ecr-image.html

Login with AWS:

```
 aws ecr get-login-password --region eu-west-1 | docker login --username AWS --password-stdin 724762920366.dkr.ecr.eu-west-1.amazonaws.com
```

Build the API:

```
docker build -t cycleforlisbon-dev-api .
```

Tag the image

```
docker tag cycleforlisbon-dev-api:latest 724762920366.dkr.ecr.eu-west-1.amazonaws.com/cycleforlisbon-dev-api:latest
```

Push the image to ECR

```
docker push 724762920366.dkr.ecr.eu-west-1.amazonaws.com/cycleforlisbon-dev-api:latest
```

### 6. Add SES DNS Records

The stack will stay in deployment until the SES identity is verified.

Go to Amazon SES grab the CNAME, MX and TXT records, and publish them to the DNS provider.

### 7. Add Certificate DNS Records

Do the same as in step 2 but now with the API Certificate

### 8. Add CNAME DNS Records

Add the CNAME records specified in the outputs of the main stack.

## Steps to deploy

### 1. Upload new docker image

See "5. Upload the docker containers to ECR" above for instructions on this

### 2. Force a new deployment on ECS

Run a `update-service --force-new-deployment` on the API service. This launches a new instance and

```
aws ecs update-service --cluster <ecs cluster> --service <ecs service> --force-new-deployment
```

### 3. Upload react files

Build the new react app, and run

```
aws s3 sync ./build s3://cycleforlisbon-dev-dashboard/ --delete
```

or

```
STACKNAME=cycleforlisbon ./aws/scripts/deploy-dashboard.sh
```

### 4. Invalidate cache on cloudfront

Invalidate the current cache. Do it only for index.html, the rest of the files should expire normally and new files should be available immediately.

```
aws cloudfront create-invalidation --distribution-id <distribution-id> --paths /index.html
```

## Notes

### SES sandbox

While an account is in sandbox mode certain restrictions apply. For example, emails can only be
sent to other verified addresses or domains.

When moving to production, see this guide to move out of sandbox mode:
https://docs.aws.amazon.com/ses/latest/dg/request-production-access.html
