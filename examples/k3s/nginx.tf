resource "kubectl_manifest" "hello" {
  yaml_body = <<YAML
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: nginx-hello
  labels:
    app: nginx-hello
spec:
  selector:
    matchLabels:
      app: nginx-hello
  template:
    metadata:
      labels:
        app: nginx-hello
    spec:
      containers:
        - image: nginxdemos/hello
          name: nginx-hello
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
YAML
}

resource "kubectl_manifest" "hello_service" {
  yaml_body = <<YAML
apiVersion: v1
kind: Service
metadata:
  name: nginx-hello
  labels:
    app: nginx-hello
spec:
  type: ClusterIP
  selector:
    app: nginx-hello
  ports:
    - port: 80
YAML
}

resource "kubectl_manifest" "hello_ingress" {
  yaml_body = <<YAML
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-hello
spec:
  rules:
    - host: "*.bnr.la"
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: nginx-hello
                port:
                  number: 80
  YAML
}
