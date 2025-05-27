set -exuo pipefail
script_dir="$(dirname "$0")"
post_out="${script_dir}/post.txt"
get_out="${script_dir}/get.txt"

wrk -c 1 -d 1s -t 1 -s bench/post_and_store.lua http://localhost:8080/url --timeout 10s > /dev/null &&
wrk -c 50 -d 30s -t 1 -s bench/post_and_store.lua http://localhost:8080/url --timeout 10s > $post_out &
wrk -c 50 -d 30s -t 1 -s bench/get_and_retrieve.lua http://localhost:8080/url -- timeout 10s > $get_out &
wait