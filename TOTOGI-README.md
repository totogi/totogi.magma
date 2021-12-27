## Requirements

1. Access to new AWS Account for isolation
2. AWS Access keys & profile setup for the account (as default profile preferrably)


## Clone Git Repository

**This step may change to reflect the modifications to magma core. 

```bash
git clone https://github.com/magma/magma.git
cd magma
git checkout v1.6.0
export MAGMA_ROOT=$(pwd)
 
git apply ../*.diff
```

## Set up vars

```bash
export YOUR_EMAIL="fernanado.medina.corey@telcodr.com"
export YOUR_ROOT_DOMAIN="totogidemo.com"
export YOUR_ORC8R_DOMAIN=*.orc8r.$YOUR_ROOT_DOMAIN
```

## SSL Certificates

**Install `certbot`**

[Get Certbot](https://eff-certbot.readthedocs.io/en/stable/install.html#alternate-installation-methods) to use locally with the `certonly` flag. 

Instructions [for Ubuntu](https://www.digitalocean.com/community/tutorials/how-to-use-certbot-standalone-mode-to-retrieve-let-s-encrypt-ssl-certificates-on-ubuntu-16-04)

Instructions for Mac:

- [Install Brew](https://brew.sh/)
- Install certbot with Brew: `brew install certbot`

Create the certificate and the private key. Requires setting ACME challenge in Route53 hosted zone. The step below generates the files `fullchain.pem` and `privkey.pem`.

```bash
certbot certonly --manual \
--preferred-challenges=dns \
--email $YOUR_EMAIL \
--server https://acme-v02.api.letsencrypt.org/directory \
--agree-tos -d $YOUR_ORC8R_DOMAIN
```

This will prompt you to deploy a DNS record for your domain with something like this:

```
Please deploy a DNS TXT record under the name:

_acme-challenge.orc8r.totogidemo.com.

with the following value:

fGaiP7x-WZjeNwu4ZhCiM_O3JS4HjSzGLMBLaPKHpNA

Before continuing, verify the TXT record has been deployed. Depending on the DNS
provider, this may take some time, from a few seconds to multiple minutes. You can
check if it has finished deploying with aid of online tools, such as the Google
Admin Toolbox: https://toolbox.googleapps.com/apps/dig/#TXT/_acme-challenge.orc8r.totogidemo.com.
Look for one or more bolded line(s) below the line ';ANSWER'. It should show the
value(s) you've just added.
```

Replace these with the values from `certbot`: 

```bash
export CHALLENGETXTDOMAIN="_acme-challenge.orc8r.totogidemo.com."
export CHALLENGETXTVALUE="fGaiP7x-WZjeNwu4ZhCiM_O3JS4HjSzGLMBLaPKHpNA"
```

First, get your hosted zone ID:

```bash
export ROOT_HOSTED_ZONE_ID_LONG=`aws route53 list-hosted-zones-by-name \
--dns-name $YOUR_ROOT_DOMAIN \
--output text \
--query 'HostedZones[0].Id'`
export ROOT_HOSTED_ZONE_ID=${ROOT_HOSTED_ZONE_ID_LONG/\/hostedzone\//}
```

Then change the update the record sets:


```bash
export R53_BATCH_CHANGE_JSON='
  {
    "Changes": [{
      "Action"              : "CREATE"
      ,"ResourceRecordSet"  : {
        "Name"              : "'"$CHALLENGETXTDOMAIN"'"
        ,"Type"             : "TXT"
        ,"TTL"              : 120
        ,"ResourceRecords"  : [{
            "Value"         : "\"'"$CHALLENGETXTVALUE"'\""
        }]
      }
    }]
  }
  '
aws route53 change-resource-record-sets \
  --hosted-zone-id $ROOT_HOSTED_ZONE_ID \
  --change-batch $R53_BATCH_CHANGE_JSON
```

Check that the results show the text record
```bash
aws route53 list-resource-record-sets \
--hosted-zone-id $ROOT_HOSTED_ZONE_ID \
--query 'ResourceRecordSets[?Type==`TXT`]'
```


## Set up Certificates:

In order to setup your `certbot` certificates and get another required LetsEncrypt certificate you can run:

```bash
mkdir -p ~/secrets/certs -p
cd ~/secrets/certs
wget https://letsencrypt.org/certs/lets-encrypt-r3.pem
sudo cp /etc/letsencrypt/live/orc8r.totogidemo.com/fullchain.pem controller.crt
sudo cp /etc/letsencrypt/live/orc8r.totogidemo.com/privkey.pem controller.key
mv lets-encrypt-r3.pem rootCA.pem
```

The earlier `certbot` command should have output certificates for you at `/etc/letsencrypt/live/orc8r.totogidemo.com/` and a full list of LetsEncrypt certificates can be found here: https://letsencrypt.org/certificates/.

Then generate the remaining certificates with the create_application_certs.sh script:

```bash
${MAGMA_ROOT}/orc8r/cloud/deploy/scripts/create_application_certs.sh $YOUR_ORC8R_DOMAIN
openssl pkcs12 -export -inkey admin_operator.key.pem -in admin_operator.pem -out admin_operator.pfx
```

Also create the 'pfx' for REST API interaction (Swagger).

```bash
openssl pkcs12 -export -inkey admin_operator.key.pem -in admin_operator.pem -out admin_operator.pfx
```

You'll need to enter a password. 

Certificate files would look like this after the procedure under ~/secrets/certs.

```
controller.crt
controller.key
rootCA.pem
bootstrapper.key
fluentd.key
fluentd.pem
certifier.key
certifier.pem
admin_operator.key.pem
admin_operator.pem
```

Note - there is no rootCA.key unless you do a self-signed version of this process.

## Preparing the Terraform deployment

Edit the terrform file located at `${MAGMA_ROOT}/orc8r/cloud/deploy/terraform/orc8r-helm-aws/examples/basic/main.tf`

- Update the `region` in both modules to the region you're deploying into
- Update `orc8r_deployment_type` to `all`
- Update `orc8r_tag` to `1.6.0` or a newer version if appropriate
- Add `orc8r_db_engine_version = "12.6"` to the `orc8r` module*** This may actually end up needing to be `"12.7"` given errors encountered when running the full `terraform apply` command below. The magma docs say 12.6 but Terraform may complain.
- Update `cluster_version` to `"1.18"`
- Update the `orc8r_domain_name` to the value of your domain name
- Add a custom `orc8r_db_password`
- Change the `orc8r_sns_email` to your email
- If an account has already been deployed to 

Comment out these variables in the orc8r-app module, so as to default to the Magma docker registry and helm repository values.
- docker_registry
- docker_user
- docker_pass
- helm_repo
- helm_user
- helm_pass

Then, deploy the root module with:

- `terraform apply -target=module.orc8r`

It should output something *like* this:

```
Apply complete! Resources: 72 added, 0 changed, 0 destroyed.

Outputs:

nameservers = tolist([
  "ns-1002.awsdns-61.net",
  "ns-1376.awsdns-44.org",
  "ns-1619.awsdns-10.co.uk",
  "ns-304.awsdns-38.com",
])
```

After it's done, copy the output Kube Config file locally:

```bash
mkdir -p ~/.kube
cp kubeconfig_orc8r ~/.kube/config
```

Then edit the following two files:
- `orc8r/cloud/deploy/terraform/orc8r-helm-aws/scripts/create_orc8r_secrets.py` 
- `orc8r/cloud/deploy/terraform/orc8r-helm-aws/secrets.tf`

By commenting out the line with: `'rootCA.key',` on it. That line is only needed for self-signed scenarios.

Then run this using terraform: 
- `terraform apply -target=module.orc8r-app.null_resource.orc8r_seed_secrets`


Then run `terraform apply`

You may also need to update the Postgres version in the following file:
- `orc8r/cloud/deploy/terraform/orc8r-helm-aws/examples/basic/main.tf`

The `orc8r_db_engine_version` value may need to be changed from `"12.6"` to `"12.7"`