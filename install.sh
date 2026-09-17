#!/bin/sh
# Installs the latest efa-cli release binary into /usr/local/bin.
# Usage: curl -fsSL https://raw.githubusercontent.com/simonjenny/efa-cli/main/install.sh | sh
set -eu

repo="simonjenny/efa-cli"
install_dir="/usr/local/bin"

os=$(uname -s)
case "$os" in
	Linux) os=linux ;;
	Darwin) os=darwin ;;
	*)
		echo "error: unsupported OS '$os'. On Windows, use install.ps1 instead." >&2
		exit 1
		;;
esac

arch=$(uname -m)
case "$arch" in
	x86_64|amd64) arch=amd64 ;;
	arm64|aarch64) arch=arm64 ;;
	*)
		echo "error: unsupported architecture '$arch'" >&2
		exit 1
		;;
esac

latest_url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest")
tag=${latest_url##*/}
if [ -z "$tag" ] || [ "$tag" = "latest" ]; then
	echo "error: could not determine the latest release tag" >&2
	exit 1
fi

asset="efa-$os-$arch"
url="https://github.com/$repo/releases/download/$tag/$asset"
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

echo "Downloading $asset ($tag)..."
curl -fsSL -o "$tmp" "$url"
chmod +x "$tmp"

if [ -w "$install_dir" ]; then
	mv "$tmp" "$install_dir/efa"
else
	echo "Installing to $install_dir requires sudo:"
	sudo mv "$tmp" "$install_dir/efa"
fi

echo "Installed $("$install_dir/efa" --version) to $install_dir/efa"
