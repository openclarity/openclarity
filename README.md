# kubei

Runtime vulnerability scanner


instructions:
1) run: 'kubectl apply -f kubei.yaml'
2) run: 'kubectl -n kubei  port-forward kubei <some port num>:8080'  (example=kubectl -n kubei  port-forward kubei 5556:8080)
3) go to 'http://localhost:<some port num>/view/' using your browser (example http://localhost:5556/view/)
4) enjoy