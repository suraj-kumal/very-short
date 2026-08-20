# Encode

The `encode` package provides a function for converting integer IDs into obfuscated Base62 strings.

## Overview

`EncodeBase62` first obfuscates the input ID using a multiplication factor and a 32-bit mask. The resulting value is then encoded using Base62.

This can be useful when IDs need to be represented as shorter, URL-friendly strings while making the original sequential ID less obvious.

## Encoding Process

The encoding consists of two steps:

1. **Obfuscate the ID**
2. **Encode the obfuscated value as Base62**

### 1. Obfuscation

The ID is multiplied by `mixMultiplier` and limited to 32 bits:

```go
func obfuscate(id int, mixMultiplier int) int {
	return (id * mixMultiplier) & mixMask
}
```

The mask is defined as:

```go
const mixMask = (1 << 32) - 1
```

This keeps only the lower 32 bits of the multiplication result.

> **Note:** The obfuscation is deterministic. The same `id` and `mixMultiplier` will always produce the same encoded value.

### 2. Base62 Encoding

The obfuscated value is converted to Base62 using the following character set:

```text
0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz
```

The character indexes are:

| Range | Characters |
| ----- | ---------- |
| 0–9   | `0-9`      |
| 10–35 | `A-Z`      |
| 36–61 | `a-z`      |

The number is repeatedly divided by `62`, with each remainder selecting a character from the Base62 alphabet.

The generated characters are then reversed to produce the final encoded string.

## API

### `EncodeBase62`

```go
func EncodeBase62(id int, mixMultiplier int) string
```

#### Parameters

* `id` — The integer ID to encode.
* `mixMultiplier` — The multiplier used during the obfuscation step.

#### Returns

A Base62-encoded string representing the obfuscated ID.

## Example

```go
encoded := EncodeBase62(12345, 31)
```

The function first calculates:

```text
obfuscated = (12345 × 31) & 0xFFFFFFFF
```

The resulting value is then converted to Base62.

Because the multiplier is part of the transformation, changing `mixMultiplier` changes the resulting encoded string.

## Zero Value

If the obfuscated value is `0`, the function returns the first Base62 character:

```text
0
```

This is handled explicitly by:

```go
if obfuscated == 0 {
	return string(chars[0])
}
```

## Important Considerations

### Deterministic Output

The encoding is deterministic:

```text
EncodeBase62(id, multiplier)
```

will always return the same value for the same inputs.

### Not Encryption

This function should **not** be considered encryption or cryptographic protection.

The transformation is simply:

```text
ID → multiplication → 32-bit mask → Base62
```

Anyone who knows the algorithm and multiplier may be able to recover or analyze the original IDs.

Use it for **obfuscation**, not for protecting sensitive information.

### Multiplier Selection

The `mixMultiplier` affects how IDs are distributed after multiplication. If reversibility or collision-free encoding is important, the multiplier should be selected carefully.

In particular, when working modulo `2^32`, a multiplier that is **odd** is invertible modulo `2^32`, while an even multiplier is not.

### Integer Size

The obfuscation step is intentionally limited to 32 bits:

```go
(id * mixMultiplier) & 0xFFFFFFFF
```

Therefore, only the lower 32 bits of the multiplication result are retained.

The behavior of very large or negative `int` values should be considered carefully if this function is used across different platforms or with IDs outside the expected range.

## Summary

`EncodeBase62` provides a compact, deterministic representation of an integer ID:

```text
        ID
         │
         ▼
   × mixMultiplier
         │
         ▼
    32-bit mask
         │
         ▼
      Base62
         │
         ▼
    Encoded string
```

It is intended for ID obfuscation and compact representation, rather than cryptographic security.

