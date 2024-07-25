import {
  vmClarityApiAxiosClient,
  vmClarityUIBackendAxiosClient,
} from "./axiosClient";
import { VMClarityUIBackendApi } from "./generated-ui-backend";
import { VMClarityApi } from "./generated-api";

const vmClarityApi = new VMClarityApi(
  undefined,
  undefined,
  vmClarityApiAxiosClient,
);
const vmClarityUIBackend = new VMClarityUIBackendApi(
  undefined,
  undefined,
  vmClarityUIBackendAxiosClient,
);

export { vmClarityApi, vmClarityUIBackend };
