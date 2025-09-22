default:
    just --list

tag := "v2.179.0-martin1"
image := "martin/supabase/auth:" + tag
destination := "containers.artifactory.schibsted.io/" + image

build:
    podman manifest create {{image}} -a
    podman build --platform linux/amd64,linux/arm64 --manifest {{image}} .
    @echo "Done! Image: {{image}}"

push:
    # podman image push {{image}}
    podman manifest push {{image}} {{destination}}
    @echo "Pushed {{destination}}"
