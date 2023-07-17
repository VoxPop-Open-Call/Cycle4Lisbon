#!/bin/bash
# Uploads a directory to cloudfront and creates an invalidation

set -e

if [[ ! -v STACKNAME ]]; then
    echo "STACKNAME is not set"
    exit 1
fi

echo "Querying AWS for stack info"
dashboard_stack_arn=$(
    aws cloudformation list-stack-resources --stack-name $STACKNAME --output json \
        | jq -r '.StackResourceSummaries[] | select(.LogicalResourceId=="DashboardCF").PhysicalResourceId'
)

dashboard_stack=$(aws cloudformation list-stack-resources --stack-name $dashboard_stack_arn --output json)
dashboard_s3=$(jq -r '.StackResourceSummaries[] | select(.LogicalResourceId=="DashboardBucket").PhysicalResourceId' <<< "$dashboard_stack")
cloudfront_id=$(jq -r '.StackResourceSummaries[] | select(.LogicalResourceId=="DashboardCFDistribution").PhysicalResourceId' <<< "$dashboard_stack")

react_path=$(pwd)/dashboard/build

echo "Building react image"
mkdir -p $react_path
DOCKER_BUILDKIT=1 docker build \
    --rm \
    --target=artifact \
    --output type=local,dest=$react_path \
    --build-arg RELEASE_NAME=$RELEASE_NAME \
    dashboard

echo "Uploading new files to s3"
aws s3 sync --delete "$react_path" "s3://$dashboard_s3/"

echo "Invalidating index.html in Cloudfront"
aws cloudfront create-invalidation --distribution-id "$cloudfront_id" --paths "/index.html"

echo "Done!"
exit 0
