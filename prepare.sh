./filebrowser config init
./filebrowser config set --address 0.0.0.0
./filebrowser config set --signup
./filebrowser config set --perm.admin
./filebrowser config set --auth.method=noauth
./filebrowser -r /mnt/k8s_artifacts/containers -R /mnt/k8s_artifacts_idc/containers
