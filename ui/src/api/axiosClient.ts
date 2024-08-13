import axios from "axios";

const openClarityApiAxiosClient = axios.create({
  baseURL: `${window.location.origin}/api`,
});

const openClarityUIBackendAxiosClient = axios.create({
  baseURL: `${window.location.origin}/ui/api`,
});

export { openClarityApiAxiosClient, openClarityUIBackendAxiosClient };
