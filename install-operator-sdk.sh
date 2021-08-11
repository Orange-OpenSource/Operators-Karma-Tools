#!/bin/bash

RELEASE_VERSION=v1.5.0

if [ -n "${HTTP_PROXY}" ]
then
  proxy="-x ${HTTP_PROXY}"
else
  proxy=""
fi

export ARCH=$(case $(arch) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(arch) ;; esac)
export OS=$(uname | awk '{print tolower($0)}')
export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}

echo "Download operator-sdk..."
curl ${proxy} -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}

echo -n "Process installation of ${RELEASE_VERSION} ?(yes/NO)"; read yn
[ "$yn" != "yes" ] && exit 0

echo "Install operator-sdk in /usr/local/bin"
chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk

exit $?


