In order to run end-to-end tests locally:

Note: Need to increase docker daemon memory to 8G. (On mac with docker desktop)
Careful, this will drain a lot from your computer cpu.

- build all images (docker build)
- replace values:

   ```shell
   sed -i 's/latest/v1.1/g' charts/kubeclarity/Chart.yaml
   sed -i 's/latest/${{ github.sha }}/g' charts/kubeclarity/values.yaml
   sed -i 's/Always/IfNotPresent/g' charts/kubeclarity/values.yaml
   ```

- make cli
- mv ./cli/bin/cli ./e2e/kubeclarity-cli
- make e2e
