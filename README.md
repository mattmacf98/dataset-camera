# Module dataset-camera

Serve images from a Viam dataset as a camera component, cycling through dataset entries on each `GetImage` call.

## Model mattmacf:dataset-camera:dataset-camera

A camera component that streams images stored in a Viam dataset. On each `GetImage` call the module fetches the next image from the dataset in chronological order, wrapping around when the end is reached. This makes it easy to replay captured data or feed dataset images into a vision pipeline.

### dataset-camera Configuration

The following attribute template can be used to configure this model:

```json
{
  "dataset_id": "your-dataset-id",
  "api_key": "your-api-key",
  "api_key_id": "your-api-key-id"
}
```

#### dataset-camera Attributes

| Name | Type | Inclusion | Description |
| --- | --- | --- | --- |
| `dataset_id` | string | Required | The ID of the Viam dataset to stream images from |
| `api_key` | string | Optional | Viam API key with access to the dataset. Falls back to the `VIAM_API_KEY` environment variable if omitted |
| `api_key_id` | string | Optional | Viam API key ID. Falls back to the `VIAM_API_KEY_ID` environment variable if omitted |

#### Authentication

Credentials can be provided in two ways:

**Option 1 — Config attributes** (recommended when the module runs on a machine where you want to scope access to a specific key):

```json
{
  "dataset_id": "your-dataset-id",
  "api_key": "your-api-key",
  "api_key_id": "your-api-key-id"
}
```

**Option 2 — Module environment variables** (useful when you want to keep credentials out of the component config, or share a key across multiple components in the same module):

```json
{
  "modules": [
    {
      "type": "registry",
      "name": "mattmacf_dataset-camera",
      "module_id": "mattmacf:dataset-camera",
      "version": "latest",
      "env": {
        "VIAM_API_KEY": "your-api-key-here",
        "VIAM_API_KEY_ID": "your-api-key-id-here"
      }
    }
  ]
}
```

The API key must have at least **read** access to the organization that owns the dataset.

#### dataset-camera Example Configuration

```json
{
  "name": "dataset-cam",
  "api": "rdk:component:camera",
  "model": "mattmacf:dataset-camera:dataset-camera",
  "attributes": {
    "dataset_id": "abcd1234-0000-0000-0000-000000000000",
    "api_key": "my-api-key",
    "api_key_id": "my-api-key-id"
  }
}
```
