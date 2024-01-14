#!/bin/bash
set -x

ANDROID_HOME=~/Android/Sdk/ ANDROID_NDK_ROOT=~/Android/Sdk/ndk/26.1.10909125/ cargo apk run -r -p boidoid
