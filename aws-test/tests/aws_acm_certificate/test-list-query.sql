select certificate_arn, domain_name, title, akas
from aws.aws_acm_certificate
where akas::text = '["{{ output.resource_aka.value }}"]'
