#!/bin/bash

tag=$1
pwd_dir=$(pwd)
tmp_dir=tmp
output="$tmp_dir/install.yaml"

if [ -z "$tag" ]; then
  echo "Usage: command image-tag"
  echo "Example: generate-install-yaml.sh v0.0.12"
  exit 1
fi

mkdir -p $tmp_dir && rm -rf $tmp_dir/config
cp -r config $tmp_dir/config

cd $tmp_dir/config/manager
kustomize edit set image controller=ghcr.io/kubesphere-sigs/ks-releaser:$tag

cd $pwd_dir
kustomize build $tmp_dir/config/default -o $output

echo "Install YAML file was generated to $output"
