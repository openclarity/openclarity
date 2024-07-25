import axios from 'axios';

const vmClarityApiAxiosClient = axios.create({
  baseURL: `${window.location.origin}/api`,
});

const vmClarityUIBackendAxiosClient = axios.create({
    baseURL: `${window.location.origin}/ui/api`,
});

export {
    vmClarityApiAxiosClient,
    vmClarityUIBackendAxiosClient,
};
