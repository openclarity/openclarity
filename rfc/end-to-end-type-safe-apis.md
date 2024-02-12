# [RFC] End to end type safe APIs

*Note: this RFC template follows HashiCrop RFC format described [here](https://works.hashicorp.com/articles/rfc-template)*

|               |                                                             |
|---------------|-------------------------------------------------------------|
| **Created**   | 2023-11-14                                                  |
| **Status**    | WIP \| InReview \| **Approved** \| Obsolete                 |
| **Owner**     | *gallotamas*                                                |
| **Approvers** | [PR-946](https://github.com/openclarity/vmclarity/pull/946) |

---

We are using OpenAPIs in the project to describe our APIs. It should be the single source of truth for all the code that we write. Automations should also be set up to make sure that changes made to the API doesn't break any other code on the backend or the UI.

## Background

The API is changing frequently and we often face the issue that something is broken (especially on the UI) because of the changes made on the API.

UI developers always have to check the open api description because they don't have typing support during development. This process is manual and error-prone. It is also very hard to maintain because you might not know all the occurrences that you have to change given a specific API change.

## Proposal

I'm focusing on the UI part here but the RFC can be extended with backend related proposals.

We could generate a client side SDK based on the OpenAPIs. The [OpenAPI generator](https://openapi-generator.tech/) could be used to generate it. There are a lot of different client side generators but let's choose the typescript-axios one. By generating TypeScript code we can ensure type safety and [axios](https://github.com/axios/axios) is a battle tested HTTP library with all the features we would ever need for making HTTP requests including interceptors. The code generation should be part of the build pipeline so whenever we make changes to the API it should be reflected in the UI build immediately and if it breaks the code then it blocks merging the PR.

Besides generating the client side code to access the APIs we should also have a strategy for caching, request deduplication, keeping data up to date. Instead of solving all of this on our own we should use a proper data-fetching library. [`TanStack Query`](https://tanstack.com/query/v5/docs/react/overview) (formerly known as `react-query`) solves all of this (and even more) for us.

The above proposal is just the groundwork required for end to end type safety. Our client side code is written in JS at the moment. We should switch to TS in all the UI code and replace the API calls with the appropriate calls using the generated SDK to ensure end to end type safety.

### Abandoned Ideas (Optional)

No abandoned ideas so far.

---

## Implementation

You can find a POC implementation on the [`open-api-codegen-poc`](https://github.com/openclarity/vmclarity/tree/open-api-codegen-poc) branch.

1. Install the openapi-generator-cli

```sh
npm install @openapitools/openapi-generator-cli -D
```


2. Run the code generation

```sh
npx @openapitools/openapi-generator-cli generate -i ../api/openapi.yaml -g typescript-axios -o ./src/api/generated --openapi-normalizer SET_TAGS_FOR_ALL_OPERATIONS=VMClarity
```

This should be added as an npm command and run in CI as well.
SET_TAGS_FOR_ALL_OPERATIONS is used to make sure that the API is properly named (without this it would be named as DefaultApi)


3. Create a single axios instance and a single instance of the API.

```ts
const axiosClient = axios.create({
    baseURL: `${window.location.origin}/api`,
});

// create an instance of the api with default configuration and the above defined axios instance.
const vmClarityApi = new VMClarityApi(undefined, undefined, axiosClient);
```


4. Use the API instance in your react-query based queries.

```ts
const { data: assets } = useQuery({
    queryKey: ['assets'],
    queryFn: () => vmClarityApi.getAssets(),
    select: (resp) => resp.data
});
```
The nice thing about the code above is that we get type safety out of the box without the need to declare the types manually. I also don't have to know the path and parameters of the specific API calls.


## UX

This RFC has no visible impacts on the UX.

## UI

This RFC has no visible impacts on the UI.
