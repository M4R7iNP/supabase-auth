default:
    just --list

tag := "v2.179.0-martin7"
image := "martin/supabase/auth:" + tag
destination := "containers.artifactory.schibsted.io/" + image

build:
    podman manifest create {{image}} -a
    podman build --jobs 4 --platform linux/amd64,linux/arm64 --build-arg RELEASE_VERSION={{tag}} --manifest {{image}} .
    @echo "Done! Image: {{image}}"

push:
    podman manifest push {{image}} {{destination}}
    @echo "Pushed {{destination}}"
