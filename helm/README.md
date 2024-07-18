
## package helm chart

    perl -pe 's/(version: )(v\d\.)(\d\.)(\d+)/$1 . $2.$3.($4 + 1)/ge' helm/Chart.yaml > helm/Chart.yaml.tmp; 
    mv helm/Chart.yaml.tmp helm/Chart.yaml

    helm package helm
    helm cm-push -u <username> -p <password> $(ls -tr -1 *.tgz | tail -n 1) <repo>

## install

    helm install <releasename> <repo>/<chartname> --namespace <namespace> --create-namespace --wait --set ingress.tls='true' --set ingress.ingessClassName='traefik' --set ingress.domain='files.example.com'

or

    helm upgrade --install <releasename> <repo>/<chartname> --namespace <namespace> --create-namespace --wait --set ingress.tls='true' --set ingress.ingessClassName='traefik' --set ingress.domain='files.example.com'

## install from local directory 

    helm upgrade --install <releasename> helm/ --namespace <namespace> --create-namespace --wait --set ingress.tls='true' --set ingress.ingessClassName='traefik' --set ingress.domain='files.example.com'

## upgrade 

    helm repo update
    helm upgrade <releasename> <repo>/<chartname> --namespace <namespace>

