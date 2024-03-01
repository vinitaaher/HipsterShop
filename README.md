# HipstersShop



## Development Guide 

This doc explains how to build and run the Hipstershop source code locally.  

### Prerequisites 
- Jaeger
- [Docker for Desktop](https://www.docker.com/products/docker-desktop).
- [Minikube](https://minikube.sigs.k8s.io/docs/start/) (optional - see Local Cluster)  or 
[Kind cluster](https://kind.sigs.k8s.io/docs/user/quick-start/) (optional - see Local Cluster)

## Steps to run on Kubernetes
1. Get the checkout of feature/develop branch .
2. Open the terminal and go to Hipstershop folder. Execute following command.
      ```sh
   minikube start
      ```
   i) Export environment variable's.
      ```sh
      export MY_IP=<Your Ip>
      export SERVICE_NAMESPACE="hipster"
      ```
   ii) If you have to send traces in Jaeger Format.
    ```sh
   envsubst < kubernetes-manifests_jaeger.yml | kubectl apply -f -
   minikube service frontend
      ```
   iii) If you have to send traces in OTLP Format.
     ```sh
      envsubst < kubernetes-manifests_otlp.yaml | kubectl apply -f -
      minikube service frontend
      ```
3. You can verify the pod,deployment and service using following command.
       
      ```sh
      kubectl get pods
      kubectl get services
      kubectl get deployment
      minikube dashboard
      ```
       
4. You can delete deployment using following command.
    ```sh
    kubectl delete -f <.yml file name>
    ``` 

## Steps to run on Kind Cluster
1. Get the checkout of feature/develop branch .
2. Open the terminal and go to Hipstershop folder. Execute following command.
      ```sh
   kind create cluster
      ```
   i) Export environment variable's.
      ```sh
      export MY_IP=<Your Ip>
      export SERVICE_NAMESPACE="hipster"
      ```
   ii) If you have to send traces in Jaeger Format.
    ```sh
   envsubst < kubernetes-manifests_jaeger.yml | kubectl apply -f -
   kubectl get services frontend
   kubectl port-forward svc/frontend 8080:80
      ```
   iii) If you have to send traces in OTLP Format.
     ```sh
      envsubst < kubernetes-manifests_otlp.yaml | kubectl apply -f -
      kubectl get services frontend
      kubectl port-forward svc/frontend 8080:80
      ```
   iv) open url: (http://localhost:8080/)
3. You can verify the pod,deployment and service using following command.

      ```sh
      kubectl get pods 
      kubectl get secrets
      kubectl get pvc
      ```
4. To start dashboard use following commands.
      ```sh
      kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.1.0/aio/deploy/recommended.yaml
      kubectl get pod -n kubernetes-dashboard
      kubectl proxy
      ```
   and open url :
   http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/#!/login
   (use your token to login.)
   

### Screenshot

 [![Screenshot of store homepage](./Dag.png)](./Dag.png)

## Steps to run on Docker

> Note : No need to run jaeger locally. It is built in docker compose.

1. Get the checkout of Docker_Compose_Addon branch
2. Open the terminal and go to Hipstershop folder. Execute following command.
  
     ```sh
     
    docker-compose up -d
     
      ```
4.  Access the web frontend through your browser 
  
  - Once run above all steps you can access frontend service at  http://localhost:8081
  
  - Access the jaeger at http://localhost:16686/ , You will see traces in jaeger
 
## Steps to run on Local Machine


### Prerequisites 
- Jaeger
- [Docker for Desktop](https://www.docker.com/products/docker-desktop).
- JDK 11
- Installation of Go
- Installation of Python
- Visual Studio


1. Currency Service (Node.js)

    ```sh
    cd currencyservice
    npm i 
    node -r ./tracing server.js tracing initialized
    
    ```
2. Cart Service (C#)
      
    ```sh
    Opent and Run the cart service in Visual Studio.
   
    ```
  
3. Payment Service(Node.js)
  
    ```sh
    cd paymentservice
    node -r ./tracing index.js  
    
    ```
    
4. Recommendation Service (Python)
  
    ```sh
    cd recommendationservice
    pip install -r requirements.txt
    opentelemetry-instrument -e none python3 recommendation_server.py
    
    ```
    
5. Shipping Service(Go)
  
    ```sh
    cd shippingservice
    go run . .
    
    ```
    
6. ProductCatlog Service (Go)
  
    ```sh
    cd productcatalogservice
    go run . .
    
    ```
    
7. Checkout Service (Go)
  
    ```sh
    cd checkoutservice
    go run . .
    
    ```
8. Email Service (Python)

    ```sh
    cd emailservice
    export OTEL_PYTHON_TRACER_PROVIDER=sdk_tracer_provider
    opentelemetry-instrument -e none python3 email_server.py
    
    ```
    
9. Ad Service (Java)


        cd adservice
        gradle build
        java -javaagent:tracinglib/opentelemetry-javaagent-all.jar \
        -Dotel.exporter=jaeger \
        -Dotel.exporter.jaeger.service.name=adService \
        -Dotel.exporter.jaeger.endpoint=localhost:14250 \
        -jar build/libs/hipstershop-0.1.0-SNAPSHOT-fat.jar
     
     
 > Note : Download opentelemetry-javaagent-all.jar : https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v0.8.0/opentelemetry-javaagent-all.jar and copy the jar file in folder adservice/tracinglib    

 10.  Access the web frontend through your browser 
  
  - Once run above all steps you can access frontend service at  http://localhost:8081
  - Start the jaeger either using binary file or using docker desktop http://localhost:16686/ , You will see traces in jaeger
    
