default:
    just --list

version := "2.188.1-martin24"
tag := "v" + version
image := "martin/supabase/auth:" + tag
destination := "containers.artifactory.schibsted.io/" + image

build:
    podman manifest create {{image}} -a
    podman build --cgroup-manager=cgroupfs --platform linux/amd64 --build-arg RELEASE_VERSION={{version}} -t {{image}}-amd64 .
    podman --connection graviton build --platform linux/arm64 --build-arg RELEASE_VERSION={{version}} -t {{image}}-arm64 .
    podman --connection graviton image save {{image}}-arm64 | podman image load
    podman manifest add --arch amd64 {{image}} containers-storage:localhost/{{image}}-amd64
    podman manifest add --arch arm64 {{image}} containers-storage:localhost/{{image}}-arm64
    @echo "Done! Image: {{image}}"

push:
    podman manifest push {{image}} {{destination}}
    @echo "Pushed {{destination}}"
