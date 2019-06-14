#!/bin/bash -e

source private/secrets.sh
export SKILL_SKIP_REQUEST_VALIDATION=true
export PORT=4443
export SKILL_USE_TLS=true
export CERT=private/certificate.pem
export KEY=private/private-key.pem
fresh
