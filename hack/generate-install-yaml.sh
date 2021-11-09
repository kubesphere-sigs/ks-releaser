#!/bin/bash

#!/bin/bash

usage() { echo "Usage: $0 -t <git tag name> [-p enable|disable]" 1>&2; exit 1; }

while getopts ":t:p:" opt; do
  case "$opt" in
    t)
      t=${OPTARG}
      ;;
    p)
      p=${OPTARG}
      [[ "$p" == "enable" || "$p" == "disable" ]] || usage
      ;;
    *)
      usage
      exit 1
      ;;
  esac
done
shift $((OPTIND-1))

if [ -z "${t}" ]; then
    usage
fi

if [ -z "${p}" ]; then
  p=enable
fi

tag=$t
pwd_dir=$(pwd)
tmp_dir=tmp
output="$tmp_dir/install.yaml"

mkdir -p $tmp_dir && rm -rf $tmp_dir/config
cp -r config $tmp_dir/config

cd $tmp_dir/config/manager
kustomize edit set image controller=ghcr.io/kubesphere-sigs/ks-releaser:$tag

if [ "$p" == "disable" ]; then
  cd $pwd_dir
  cd $tmp_dir/config/default
  kustomize edit remove resource ../prometheus
  output="$tmp_dir/install-no-monitor.yaml"
fi

cd $pwd_dir
kustomize build $tmp_dir/config/default -o $output

echo "Install YAML file was generated to $output"
