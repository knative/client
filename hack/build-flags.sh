function build_flags() {
 local now="$(date -u '+%Y-%m-%d %H:%M:%S')"
  local rev="$(git rev-parse --short HEAD)"
  local pkg="github.com/knative/client/pkg/kn/commands"
  local version="${TAG:-}"
  # Use vYYYYMMDD-local-<hash> for the version string, if not passed.
  if [[ -z "${version}" ]]; then
    # Get the commit, excluding any tags but keeping the "dirty" flag
    local commit="$(git describe --always --dirty --match '^$')"
    [[ -n "${commit}" ]] || abort "error getting the current commit"
    version="v$(date +%Y%m%d)-local-${commit}"
  fi
  echo "-X '${pkg}.BuildDate=${now}' -X ${pkg}.Version=${version} -X ${pkg}.GitRevision=${rev}"
}
