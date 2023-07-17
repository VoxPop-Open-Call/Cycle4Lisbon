#!/bin/bash
# Builds and uploads docker images to ECR

set -e

if [[ ! -v STACKNAME ]]; then
    echo "STACKNAME is not set"
    exit 1
fi

# Set a default for the AWS_DEFAULT_REGION if not passed in
AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:=eu-west-1}"


echo "Querying AWS for stack info"
api_stack_arn=$(
    aws cloudformation list-stack-resources --stack-name $STACKNAME --output json \
        | jq -r '.StackResourceSummaries[] | select(.LogicalResourceId=="API").PhysicalResourceId'
)

api_stack=$(aws cloudformation list-stack-resources --stack-name "$api_stack_arn" --output json)
cluster_name=$(jq -r '.StackResourceSummaries[] | select(.LogicalResourceId=="ECSCluster").PhysicalResourceId' <<< "$api_stack")
service_name=$(jq -r '.StackResourceSummaries[] | select(.LogicalResourceId=="ECSService").PhysicalResourceId' <<< "$api_stack")

ecr_name=$(jq -r '.StackResourceSummaries[] | select(.LogicalResourceId=="ECRRepository").PhysicalResourceId' <<< "$api_stack")
ecr_uri=$(
    aws ecr describe-repositories --repository-names "$ecr_name" --output json \
        | jq -r '.repositories[0].repositoryUri'
)


echo "Building API Image"
docker build -f api/Dockerfile -t "$ecr_uri:latest" ./api/

echo "Authenticating with"
aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login --username AWS --password-stdin $ecr_uri

echo "Pushing Images"
docker push "$ecr_uri:latest"

echo "Forcing ECS Deployment"
aws ecs update-service --cluster "$cluster_name" --service "$service_name" --force-new-deployment &> /dev/null

