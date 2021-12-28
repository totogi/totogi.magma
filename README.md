## This Repo

**TODOs:**

0. Review alternate process using cloudstrapper to avoid much of this overhead and provide an SSH entrypoint
1. Change clone instructions to this repo
2. Remove steps in Readme that simply using this repo avoids
3. Test cleaned/modified terraform template in new AWS account and remove those custom steps  

## Prerequisites and Required Software

- [Install The AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- Obtain and configure AWS CLI credentials for the AWS Account you are deploying to

- [Install `jq`](https://stedolan.github.io/jq/download/)
- [Install `kubectl`](https://kubernetes.io/docs/tasks/tools/)
- [Install `aws-iam-authenticator`](https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html)

## Clone Git Repository

**This step may change to reflect the modifications to magma core or this repo

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

In each file, comment out the line with: `'rootCA.key',` on it. That line is only needed for self-signed scenarios.

Then run this using terraform: 
- `terraform apply -target=module.orc8r-app.null_resource.orc8r_seed_secrets`

Then run `terraform apply`

You may also need to update the Postgres version in the following file:
- `orc8r/cloud/deploy/terraform/orc8r-helm-aws/examples/basic/main.tf`

The `orc8r_db_engine_version` value may need to be changed from `"12.6"` to `"12.7"`

Check on pods: 
kubectl get pods --all-namespaces



# Pod Orchestration

## Create the Orchestrator admin user:

```bash
export ORC_POD=$(kubectl --namespace orc8r get pod -l app.kubernetes.io/component=orchestrator -o jsonpath='{.items[0].metadata.name}')
kubectl --namespace orc8r exec ${ORC_POD} -- /var/opt/magma/bin/accessc add-existing -admin -cert /var/opt/magma/certs/admin_operator.pem admin_operator
```

Verify it was created:

```bash
kubectl --namespace orc8r exec ${ORC_POD} -- /var/opt/magma/bin/accessc list-certs
```

## Create an NMS user for the ‘master’ organization

Get the name of the NMS Pod and set it as an env variable:

```bash
export NMS_POD=$(kubectl --namespace orc8r get pod -l  app.kubernetes.io/component=magmalte -o jsonpath='{.items[0].metadata.name}')
```

Replace `ADMIN_USER_EMAIL` and `ADMIN_USER_PASSWORD` values in the command below with user and password values to be used.

```bash
kubectl --namespace orc8r exec -it ${NMS_POD} -- yarn setAdminPassword master ADMIN_USER_EMAIL ADMIN_USER_PASSWORD
```


# DNS Resolution

Assuming that the Route 53 Hosted Zone for the main domain (e.g. totogidemo.com) already exists, we need to add the name servers for the newly created subdomain (orc8r.totogidemo.com) in this example to the already existing Route 53 hosted zone.

Navigate to the same directory as the `main.tf` file: `${MAGMA_ROOT}/orc8r/cloud/deploy/terraform/orc8r-helm-aws/examples/basic/main.tf`

Then run: `terraform output`

You should see something like this:

```js
nameservers = tolist([
  "ns-1405.awsdns-47.org",
  "ns-1746.awsdns-26.co.uk",
  "ns-201.awsdns-25.com",
  "ns-695.awsdns-22.net",
])
```

Next, add the values you see to the main Route 53 main hosted zone (e.g. totogidemo.com):

```bash
export TF_NAMESERVERS=$(terraform output -json nameservers)

```
**TODO:**  
- Use JQ to parse the output of the command above and create the required JSON for the R53 API call below
- Update ROOT_HOSTED_ZONE_ID below to the relevant subdomain


"ResourceRecords": [
  {"Value": "ns-915.awsdns-50.net"},
  {"Value": "ns-1814.awsdns-34.co.uk"},
  {"Value": "ns-1271.awsdns-30.org"},
  {"Value": "ns-213.awsdns-26.com"}
]


```bash
export R53_BATCH_CHANGE_JSON='
  {
    "Changes": [{
      "Action"              : "CREATE"
      ,"ResourceRecordSet"  : {
        "Name"              : "'"$CHALLENGETXTDOMAIN"'"
        ,"Type"             : "NS"
        ,"TTL"              : 120
        ,"ResourceRecords"  : [{
            '"$DNS_RECORDS"'
        }]
      }
    }]
  }
  '
aws route53 change-resource-record-sets \
  --hosted-zone-id $ROOT_HOSTED_ZONE_ID \
  --change-batch $R53_BATCH_CHANGE_JSON
```


Check that the results show the NS records
```bash
aws route53 list-resource-record-sets \
--hosted-zone-id $ROOT_HOSTED_ZONE_ID \
--query 'ResourceRecordSets[?Type==`NS`]'
```


After this point, NMS UI can be reached via: https://master.nms.orc8r.totogidemo.com/

For accessing the Swagger UI, use Firefox, add “admin_operator.pfx”, then navigate to https://api.orc8r.totogidemo.com/swagger/v1/ui/ 
Should look like below.

**TODO** Add link to the pfx process

## Connecting to Totogi Charging System

**TODOs:** 
- Consolidate steps from these documents:
- https://docs.google.com/document/d/1j54886bN_kGzyW0039gg3vs2aPEvy7ZW55EHW0qORuI/edit#heading=h.2cy6sykle8uf
- https://docs.google.com/document/d/1bxfXMg6BF7zp8mBqU2RFo433eqH1SLpucO77iaRfkmE/edit?hl=en&forcehl=1#
- Create the totogi tenant, get the user credentials, etc. 
- Attempt to automate this process for us if possible
- Use the FedGW CLI to mock data coming from Magma (not end to end, just mocked diamter results)

## Testing with srsRAn

**TODOs:**
- Setup srsRAN for testing end to end (without eSIMs/UE Devices)
- https://docs.srsran.com/en/latest/
- Example: https://www.youtube.com/watch?v=ipiW9JXXiso&ab_channel=WavelabsTechnologies




