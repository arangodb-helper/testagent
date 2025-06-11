# Making pull request
Before submitting a pull request, make sure that all tests pass. To run the tests run `make tests`.

# How to release testagent
To release a new version of testagent, simply checkout the latest commit in the `master` branch and run one of the following make targets: `release-patch`, `release-minor` or `release-major`.  
This will do the following:
* Bump the version in the `VERSION` file according to the selected release type.
* Build new docker image
* Tag the new docker image with version tag and `latest` tag and push it
* Tag the release commit with the version tag
* Bump the version in the `VERSION` file again with the `+git` extension
  
This described process implies that releases are made only from `master` branch and therefore every release is the latest version at the time of its making, which is expressed with the `latest` tag on the docker image. If you wish to support several branches simultaneously, make target `release-patch-legacy` can be used for that. It will make a patch release and push the new image to the docker registry, but it will not tag it as `latest`.
