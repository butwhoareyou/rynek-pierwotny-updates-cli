cd ../dist || exit
AWS_ACCESS_KEY=root AWS_SECRET_KEY=password ./cli offers-updates \
--aws.region="eu-west-1" \
--aws.s3.bucket="offer-updates-1" \
--aws.endpoint="http://localhost:9000" \
--request.regions=120 \
--url="https://rynekpierwotny.pl" \
--api-url="https://rynekpierwotny.pl/api" \
--debug