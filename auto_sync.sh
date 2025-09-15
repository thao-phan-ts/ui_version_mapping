#! /usr/bin/env bash

rootdir=./submodules
de_base_dir="de_base"
if [ ! -d $rootdir ]; then
  mkdir $rootdir
fi

if [ "$1" -eq "1" ]; then
  cd $rootdir
  repos=("digital_journey" "decision_engine" "platform_integration")
  checkoutDirs=("migration" "etc/production migration" "etc piconfig")
  length=${#repos[@]}
  # Loop through the array
  for ((i=0; i<$length; i++))
  do
     echo "---------------"
     repo="${repos[$i]}"
     checkoutDir="${checkoutDirs[$i]}"
     echo "sparse checkout $repo with checkoutDir $checkoutDir"
     rm -rf $repo
     git init
     git clone --filter=blob:none --no-checkout --depth 1 --sparse git@github.com:tsocial/$repo.git
     cd $repo
     git config core.sparseCheckout true
     echo '' >> .git/info/sparse-checkout
     git sparse-checkout add $checkoutDir
     git checkout
     git pull origin master
     cd ..
  done
  cd ..
else
  for dir in $rootdir/*
  do
    echo "---------------"
    # ignore update de_base dir
    if [[ $dir == *"$de_base_dir"* ]]; then
      echo "skip $de_base_dir"
      continue
    fi
    cd $dir
    git status >/dev/null 2>&1
    # check if exit status of above was 0, indicating we're in a git repo
    [ $(echo $?) -eq 0 ] && echo "updating ${dir%*/}..." && git pull origin master
    cd ../..
  done
fi

# sparse checkout repo decision_engine to have base commit
repo="decision_engine"
# tag: v1.196.0
base_commit="d623cc45cb3e7201fa0946342cbc5d046ddd8d38"
echo "---------------"
cd $rootdir
if [ ! -d $de_base_dir ]; then
  mkdir $de_base_dir
  echo "init sparse checkout repo: $repo... with base commit $base_commit"
  git init
  git clone --filter=blob:none --no-checkout --sparse git@github.com:tsocial/$repo.git $de_base_dir
  cd $de_base_dir
  git config core.sparseCheckout true
  echo '' >> .git/info/sparse-checkout
  git sparse-checkout add etc/production migration
  echo "checkout $base_commit"
  git checkout $base_commit
else
  echo "existed $de_base_dir, updating..."
  cd $de_base_dir
  git pull origin master
  echo "checkout $base_commit"
  git checkout $base_commit
fi

