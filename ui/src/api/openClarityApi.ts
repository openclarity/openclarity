import {
  openClarityApiAxiosClient,
  openClarityUIBackendAxiosClient,
} from "./axiosClient";
import { OpenClarityUIBackendApi } from "./generated-ui-backend";
import { OpenClarityApi } from "./generated-api";

const openClarityApi = new OpenClarityApi(
  undefined,
  undefined,
  openClarityApiAxiosClient,
);
const openClarityUIBackend = new OpenClarityUIBackendApi(
  undefined,
  undefined,
  openClarityUIBackendAxiosClient,
);

export { openClarityApi, openClarityUIBackend };
