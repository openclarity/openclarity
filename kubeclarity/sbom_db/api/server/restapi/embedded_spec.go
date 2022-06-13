// Code generated by go-swagger; DO NOT EDIT.

package restapi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
)

var (
	// SwaggerJSON embedded version of the swagger document used at generation time
	SwaggerJSON json.RawMessage
	// FlatSwaggerJSON embedded flattened version of the swagger document used at generation time
	FlatSwaggerJSON json.RawMessage
)

func init() {
	SwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "title": "KubeClarity SBOM DB APIs",
    "version": "1.0.0"
  },
  "basePath": "/api",
  "paths": {
    "/sbomDB/{resourceHash}": {
      "get": {
        "summary": "Get an SBOM from DB by resource hash.",
        "parameters": [
          {
            "$ref": "#/parameters/resourceHash"
          }
        ],
        "responses": {
          "200": {
            "description": "Success",
            "schema": {
              "$ref": "#/definitions/SBOM"
            }
          },
          "404": {
            "description": "SBOM not found."
          },
          "default": {
            "$ref": "#/responses/UnknownError"
          }
        }
      },
      "put": {
        "summary": "Store an SBOM in DB for the given resource hash.",
        "parameters": [
          {
            "$ref": "#/parameters/resourceHash"
          },
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/SBOM"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "SBOM created in DB.",
            "schema": {
              "$ref": "#/definitions/SuccessResponse"
            }
          },
          "default": {
            "$ref": "#/responses/UnknownError"
          }
        }
      }
    }
  },
  "definitions": {
    "ApiResponse": {
      "description": "An object that is returned in all cases of failures.",
      "type": "object",
      "properties": {
        "message": {
          "type": "string"
        }
      }
    },
    "SBOM": {
      "description": "Software Bill Of Materials as stored in the DB.",
      "type": "object",
      "properties": {
        "sbom": {
          "type": "string",
          "format": "byte"
        }
      }
    },
    "SuccessResponse": {
      "description": "An object that is returned in cases of success that returns nothing.",
      "type": "object",
      "properties": {
        "message": {
          "type": "string"
        }
      }
    }
  },
  "parameters": {
    "resourceHash": {
      "type": "string",
      "name": "resourceHash",
      "in": "path",
      "required": true
    }
  },
  "responses": {
    "Success": {
      "description": "Success message",
      "schema": {
        "$ref": "#/definitions/SuccessResponse"
      }
    },
    "UnknownError": {
      "description": "Unknown error",
      "schema": {
        "$ref": "#/definitions/ApiResponse"
      }
    }
  }
}`))
	FlatSwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "title": "KubeClarity SBOM DB APIs",
    "version": "1.0.0"
  },
  "basePath": "/api",
  "paths": {
    "/sbomDB/{resourceHash}": {
      "get": {
        "summary": "Get an SBOM from DB by resource hash.",
        "parameters": [
          {
            "type": "string",
            "name": "resourceHash",
            "in": "path",
            "required": true
          }
        ],
        "responses": {
          "200": {
            "description": "Success",
            "schema": {
              "$ref": "#/definitions/SBOM"
            }
          },
          "404": {
            "description": "SBOM not found."
          },
          "default": {
            "description": "Unknown error",
            "schema": {
              "$ref": "#/definitions/ApiResponse"
            }
          }
        }
      },
      "put": {
        "summary": "Store an SBOM in DB for the given resource hash.",
        "parameters": [
          {
            "type": "string",
            "name": "resourceHash",
            "in": "path",
            "required": true
          },
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/SBOM"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "SBOM created in DB.",
            "schema": {
              "$ref": "#/definitions/SuccessResponse"
            }
          },
          "default": {
            "description": "Unknown error",
            "schema": {
              "$ref": "#/definitions/ApiResponse"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "ApiResponse": {
      "description": "An object that is returned in all cases of failures.",
      "type": "object",
      "properties": {
        "message": {
          "type": "string"
        }
      }
    },
    "SBOM": {
      "description": "Software Bill Of Materials as stored in the DB.",
      "type": "object",
      "properties": {
        "sbom": {
          "type": "string",
          "format": "byte"
        }
      }
    },
    "SuccessResponse": {
      "description": "An object that is returned in cases of success that returns nothing.",
      "type": "object",
      "properties": {
        "message": {
          "type": "string"
        }
      }
    }
  },
  "parameters": {
    "resourceHash": {
      "type": "string",
      "name": "resourceHash",
      "in": "path",
      "required": true
    }
  },
  "responses": {
    "Success": {
      "description": "Success message",
      "schema": {
        "$ref": "#/definitions/SuccessResponse"
      }
    },
    "UnknownError": {
      "description": "Unknown error",
      "schema": {
        "$ref": "#/definitions/ApiResponse"
      }
    }
  }
}`))
}