#!/usr/bin/env bash
#
#           File Manager Installer Script
#
#   GitHub: https://github.com/hacdias/filemanager
#   Issues: https://github.com/hacdias/filemanager/issues
#   Requires: bash, mv, rm, tr, type, grep, sed, curl/wget
#
#   This script installs File Manager to your path.
#   Usage:
#
#   $ curl -fsSL https://henriquedias.com/filemanager/get.sh | bash
#    or
#   $ wget -qO- https://henriquedias.com/filemanager/get.sh | bash
#
#   In automated environments, you may want to run as root.
#   If using curl, we recommend using the -fsSL flags.
#
#   This should work on Mac, Linux, and BSD systems, and
#   hopefully Windows with Cygwin. Please open an issue if
#   you notice any bugs.
#

install_filemanager()
{
    trap 'echo -e "Aborted, error $? in command: $BASH_COMMAND"; trap ERR; return 1' ERR
    filemanager_os="unsupported"
    filemanager_arch="unknown"
    install_path="/usr/local/bin"

    # Termux on Android has $PREFIX set which already ends with /usr
    if [[ -n "$ANDROID_ROOT" && -n "$PREFIX" ]]; then
        install_path="$PREFIX/bin"
    fi

    # Fall back to /usr/bin if necessary
    if [[ ! -d $install_path ]]; then
        install_path="/usr/bin"
    fi

    # Not every platform has or needs sudo (https://termux.com/linux.html)
    ((EUID)) && [[ -z "$ANDROID_ROOT" ]] && sudo_cmd="sudo"

    #########################
    # Which OS and version? #
    #########################

    filemanager_bin="filemanager"

    # NOTE: `uname -m` is more accurate and universal than `arch`
    # See https://en.wikipedia.org/wiki/Uname
    unamem="$(uname -m)"
    if [[ $unamem == *aarch64* ]]; then
        filemanager_arch="arm"
    elif [[ $unamem == *64* ]]; then
        filemanager_arch="amd64"
    elif [[ $unamem == *86* ]]; then
        filemanager_arch="386"
    elif [[ $unamem == *arm* ]]; then
        filemanager_arch="arm"
    else
        echo "Aborted, unsupported or unknown architecture: $unamem"
        return 2
    fi

    unameu="$(tr '[:lower:]' '[:upper:]' <<<$(uname))"
    if [[ $unameu == *DARWIN* ]]; then
        filemanager_os="darwin"
    elif [[ $unameu == *LINUX* ]]; then
        filemanager_os="linux"
    elif [[ $unameu == *FREEBSD* ]]; then
        filemanager_os="freebsd"
    elif [[ $unameu == *NETBSD* ]]; then
        filemanager_os="netbsd"
    elif [[ $unameu == *OPENBSD* ]]; then
        filemanager_os="openbsd"
    elif [[ $unameu == *WIN* || $unameu == MSYS* ]]; then
        # Should catch cygwin
        sudo_cmd=""
        filemanager_os="windows"
        filemanager_dl_ext=".exe"
    else
        echo "Aborted, unsupported or unknown OS: $uname"
        return 6
    fi

    ########################
    # Download and extract #
    ########################

    echo "Downloading File Manager for $filemanager_os/$filemanager_arch..."
    filemanager_file="${filemanager_os}-$filemanager_arch-filemanager$filemanager_dl_ext"
    filemanager_tag="$(curl -s https://api.github.com/repos/hacdias/filemanager/releases/latest | grep -o '"tag_name": ".*"' | sed 's/"//g' | sed 's/tag_name: //g')"
    filemanager_url="https://github.com/hacdias/filemanager/releases/download/$filemanager_tag/$filemanager_file"
    echo "$filemanager_url"

    # Use $PREFIX for compatibility with Termux on Android
    rm -rf "$PREFIX/tmp/$filemanager_bin"

    if type -p curl >/dev/null 2>&1; then
        curl -fsSL "$filemanager_url" -o "$PREFIX/tmp/$filemanager_bin"
    elif type -p wget >/dev/null 2>&1; then
        wget --quiet "$filemanager_url" -O "$PREFIX/tmp/$filemanager_bin"
    else
        echo "Aborted, could not find curl or wget"
        return 7
    fi

    chmod +x "$PREFIX/tmp/$filemanager_bin"

    echo "Putting filemanager in $install_path (may require password)"
    $sudo_cmd mv "$PREFIX/tmp/$filemanager_bin" "$install_path/$filemanager_bin"
    if setcap_cmd=$(PATH+=$PATH:/sbin type -p setcap); then
        $sudo_cmd $setcap_cmd cap_net_bind_service=+ep "$install_path/$filemanager_bin"
    fi

    if type -p $filemanager_bin >/dev/null 2>&1; then
        echo "Successfully installed"
        trap ERR
        return 0
    else
        echo "Something went wrong, File Manager is not in your path"
        trap ERR
        return 1
    fi
}

install_filemanager
