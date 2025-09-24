default:
    just --list

tag := "v2.179.0-martin16"
image := "martin/supabase/auth:" + tag
destination := "containers.artifactory.schibsted.io/" + image

build:
    podman manifest create {{image}} -a
    podman build --platform linux/amd64 --build-arg RELEASE_VERSION={{tag}} -t {{image}}-amd64 .
    podman --connection graviton build --platform linux/arm64 --build-arg RELEASE_VERSION={{tag}} -t {{image}}-arm64 .
    podman --connection graviton image save {{image}}-arm64 | podman image load
    podman manifest add --arch amd64 {{image}} containers-storage:localhost/{{image}}-amd64
    podman manifest add --arch arm64 {{image}} containers-storage:localhost/{{image}}-arm64
    @echo "Done! Image: {{image}}"

push:
    podman manifest push {{image}} {{destination}}
    @echo "Pushed {{destination}}"
