select title, akas, region, account_id
from aws.aws_sns_topic
where topic_arn = '{{ output.resource_aka.value }}'