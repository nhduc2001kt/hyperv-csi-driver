docker build -t nhduc2001kt/hyperv-csi-driver:0.1.0 .
docker push nhduc2001kt/hyperv-csi-driver:0.1.0
docker run --network talos-default -d --restart=always -p 127.0.0.1:2375:2375 -v /var/run/docker.sock:/var/run/docker.sock alpine/socat tcp-listen:2375,fork,reuseaddr unix-connect:/var/run/docker.sock

docker run --rm -it \
  --name tutorial \
  --hostname talos-cp \
  --read-only \
  --privileged \
  --security-opt seccomp=unconfined \
  --mount type=tmpfs,destination=/run \
  --mount type=tmpfs,destination=/system \
  --mount type=tmpfs,destination=/tmp \
  --mount type=volume,destination=/system/state \
  --mount type=volume,destination=/var \
  --mount type=volume,destination=/etc/cni \
  --mount type=volume,destination=/etc/kubernetes \
  --mount type=volume,destination=/usr/libexec/kubernetes \
  --mount type=volume,destination=/opt \
  -e PLATFORM=container \
  --network talos-default \
  ghcr.io/siderolabs/talos:v1.9.2