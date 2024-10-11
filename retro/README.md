# retro

`retro` is a golib Golang sub-package designed to support backward-compatible models through semantic versioning. It introduces a new struct field tag, `retro`, which manages the serialization and deserialization of model fields based on specified versions.

`retro` works with `json`, `yaml`, and `toml` formats and dynamically marshals or unmarshals model data according to the version in use.

## Key Features

- The `retro:"default"` tag indicates which fields should be considered when no version is specified.
- Fields without the `retro` tag but with format tags (like `json`, `yaml`, or `toml`), and without `omitempty`, will be included in all versions.
- `retro` supports logical operators like `>`, `<`, `>=`, and `<=` in version constraints, allowing fine-grained control over when fields should be included or excluded based on the version.
- `retro` intelligently handles conflicting version definitions, such as `retro:">v1.0.0,>v1.0.3"`, ensuring valid behavior and in case on conflicts or any wrong retro tag definition the field will be ignored in the serialization and deserialization.
- Versioning exceptions can be managed as well. For example, `retro:">v1.0.0,v0.0.3"` will include the field even if `v0.0.3` doesn’t meet the condition `>v1.0.0` due to the explicit exception.
- `retro` allows easily to activate the standard serialization deserialization for your model and to work only with the standard methodologies if needed bypassing `retro` features.

Additionally, `retro` respects custom marshal/unmarshal behavior defined for fields in the `json`, `yaml`, and `toml` formats, ensuring seamless integration with your model's existing logic.


## Examples and use cases

The retro_test.go file contains real world scenarios and examples on the retro features and usage.  

