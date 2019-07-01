#!/bin/bash -ex

cd $(dirname $0)/..

# if [[ -n $(git status -s) ]]; then
#     echo "Working directory not clean"
#     exit 1
# fi

# . private/s3-credentials.sh
# TABLE_NAME_OVERRIDE=TestAlexaTwitterRequests ginkgo -r

export SHA=$(git rev-parse --short HEAD)
export APP_NAME=alexa-twitter-$SHA

for region in 'eu-de'; do
    open https://login.$region.bluemix.net/UAALoginServerWAR/passcode

    cf login -a api.$region.bluemix.net --sso -o $(lpass show Personal\\api_keys/Alexa-Twitter-Skill --notes) -s alexa

    cf push --no-start $APP_NAME --hostname alexa-twitter
    cf set-env $APP_NAME APPLICATION_ID $(lpass show Personal\\api_keys/Alexa-Twitter-Skill --password)
    cf set-env $APP_NAME CONSUMER_KEY $(lpass show Personal\\api_keys/Twitter-Skill-Consumer-Key --username)
    cf set-env $APP_NAME CONSUMER_SECRET $(lpass show Personal\\api_keys/Twitter-Skill-Consumer-Key --password)
    # cf set-env $APP_NAME ACCESS_KEY_ID $(lpass show Personal\\api_keys/Alexa-Twitter-AWS --username)
    # cf set-env $APP_NAME SECRET_ACCESS_KEY $(lpass show Personal\\api_keys/Alexa-Twitter-AWS --password)
    cf restart $APP_NAME

    export OLD_RELEASES=$(cf apps | grep alexa-twitter | grep -v $SHA | cut -f 1 -d ' ')

    for release in $OLD_RELEASES; do
        cf delete -f $release
    done
done
