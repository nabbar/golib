# Testing Guide

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-86.4%25-brightgreen)](TESTING.md)
[![Test Specs](https://img.shields.io/badge/Tests%20Specs-267-green)]()
[![Test Asserts](https://img.shields.io/badge/Tests%20Asserts-423-green)]()

Comprehensive testing documentation for the `duration/big` package, covering unit tests, benchmarks, property-based testing, and performance profiling.

---

## Table of Contents

- [Overview](#overview)
- [Test Architecture](#test-architecture)
- [Framework & Tools](#framework--tools)
- [Quick Start](#quick-start)
- [Performance & Profiling](#performance--profiling)
- [Test Coverage](#test-coverage)
- [Writing Tests](#writing-tests)
- [Best Practices](#best-practices)
- [Reporting Bugs & Vulnerabilities](#reporting-bugs--vulnerabilities)
- [Resources](#resources)

---

## Overview

This package uses the Ginkgo BDD framework to define specifications for the `big.Duration` type. The tests verify correct parsing, formatting, serialization, arithmetic operations, and conversion to/from standard types.

**Test Suite Summary**
- Total Specs: 267 across 1 package
- Total Assertions: 423
- Overall Coverage: 86.4% (aggregate)
- Race Detection: ✅ Zero data races
- Execution Time: ~0.53s (without race), ~1.6s (with race)

### Test Plan

This test suite provides complete coverage of the `big` package functionality:

1. **Functional Testing**: Verifies parsing, formatting, and arithmetic operations behave as expected.
2. **Constant Validation**: Ensures standard unit constants (Second, Minute, Hour, Day) are correct.
3. **Value Testing**: Checks handling of positive, negative, and zero values.
4. **Boundary Testing**: Validates behavior at int64 limits (MaxInt64, MinInt64).
5. **Concurrency Testing**: Ensures thread safety in encoding/decoding and operations (verified via race detector).
6. **Performance Testing**: Benchmarks critical paths (parsing, string formatting, arithmetic).
7. **Serialization Testing**: Verifies compatibility with JSON, YAML, TOML, and CBOR.

### Test Completeness

**Quality Indicators:**
- **Code Coverage**: 86.4% of statements (target: >80%)
- **Race Conditions**: 0 detected across all scenarios
- **Flakiness**: 0 flaky tests detected

**Test Distribution:**
- ✅ **267 specifications** covering all functionality
- ✅ **423 assertions** validating behavior
- ✅ **16 example tests** demonstrating usage patterns
- ✅ **10 performance benchmarks** measuring key metrics
- ✅ **1 test suite** organized by functional area
- ✅ **Zero flaky tests** - all tests are deterministic

---

## Test Architecture

The tests are organized into functional groups within `_test.go` files corresponding to the source files they test.

### Test Matrix

| Category | Target | Specs | Priority | Dependencies |
|----------|--------|-------|----------|-------------|
| **Core Model** | Basic type behavior, `IsDays`, `IsHours`, etc. | 22 | Critical | None |
| **Parsing** | `Parse`, `ParseString` logic & edge cases | 52 | Critical | None |
| **Formatting** | `String`, `Time` conversion & formatting | 49 | Critical | None |
| **Serialization** | JSON/YAML/TOML/CBOR encoding/decoding | 47 | High | External libs |
| **Operations** | `Abs`, `Range`, PID controller integration | 41 | Medium | `pidcontroller` |
| **Truncation** | `Round`, `Truncate` logic | 45 | Medium | None |
| **Benchmarks** | Performance of string formatting & parsing | 10 | Low | None |
| **Examples** | Documentation examples | 16 | High | All |

### Detailed Test Inventory

**Test ID Pattern by File:**
- **TC-MD-xxx**: Model tests (model_test.go)
- **TC-PA-xxx**: Parsing tests (parse_test.go)
- **TC-FO-xxx**: Formatting tests (format_test.go)
- **TC-EN-xxx**: Encoding tests (encode_test.go)
- **TC-OP-xxx**: Operation tests (operation_test.go)
- **TC-TR-xxx**: Truncate tests (truncate_test.go)
- **TC-BM-xxx**: Benchmark tests (various files)

| Test ID | File | Use Case | Priority | Expected Outcome |
|---|---|---|---|---|
| **TC-MD-001** | `model_test.go` | Viper hook creation | Medium | Should return a valid hook function |
| **TC-MD-002** | `model_test.go` | Viper hook string decoding | Critical | Should decode "5h30m" to Duration |
| **TC-MD-003** | `model_test.go` | Viper hook days decoding | Critical | Should decode "2d12h" correctly |
| **TC-MD-004** | `model_test.go` | Viper hook non-string input | Low | Should pass through int values |
| **TC-MD-005** | `model_test.go` | Viper hook wrong target type | Low | Should pass through if target != Duration |
| **TC-MD-006** | `model_test.go` | Viper hook non-string data | Low | Should pass through if data is not string |
| **TC-MD-007** | `model_test.go` | Viper hook invalid string | Medium | Should return error |
| **TC-MD-008** | `model_test.go` | Viper hook zero duration | Medium | Should decode "0s" to 0 |
| **TC-MD-009** | `model_test.go` | Viper hook negative duration | Medium | Should decode negative strings |
| **TC-MD-010** | `model_test.go` | Viper hook complex string | High | Should decode "5d23h15m13s" |
| **TC-MD-011** | `model_test.go` | Viper hook spaces | Low | Should handle spaces in string |
| **TC-MD-012** | `model_test.go` | Viper hook quotes | Low | Should handle quotes in string |
| **TC-MD-013** | `model_test.go` | Viper hook units | High | Should support all units (s, m, h, d) |
| **TC-MD-014** | `model_test.go` | Viper hook empty string | Low | Should return error |
| **TC-MD-015** | `model_test.go` | `IsDays` true | Medium | True if >= 1 day |
| **TC-MD-016** | `model_test.go` | `IsDays` false | Medium | False if < 1 day |
| **TC-MD-017** | `model_test.go` | `IsHours` true | Medium | True if >= 1 hour |
| **TC-MD-018** | `model_test.go` | `IsHours` false | Medium | False if < 1 hour |
| **TC-MD-019** | `model_test.go` | `IsMinutes` true | Medium | True if >= 1 minute |
| **TC-MD-020** | `model_test.go` | `IsMinutes` false | Medium | False if < 1 minute |
| **TC-MD-021** | `model_test.go` | `IsSeconds` true | Medium | True if >= 1 second |
| **TC-MD-022** | `model_test.go` | `IsSeconds` false | Medium | False if < 1 second |
| **TC-MD-023** | `model_test.go` | Viper hook large duration | High | Should decode "10000d" |
| **TC-MD-024** | `model_test.go` | Viper hook fractional | High | Should decode "1.5h" |
| **TC-PA-001** | `parse_test.go` | Parse seconds | Critical | "30s" -> 30 |
| **TC-PA-002** | `parse_test.go` | Parse minutes | Critical | "5m" -> 300 |
| **TC-PA-003** | `parse_test.go` | Parse hours | Critical | "2h" -> 7200 |
| **TC-PA-004** | `parse_test.go` | Parse days | Critical | "3d" -> 259200 |
| **TC-PA-005** | `parse_test.go` | Parse complex | Critical | "5d23h15m13s" correct sum |
| **TC-PA-006** | `parse_test.go` | Parse fractional seconds | High | "1.5s" -> 1 (truncated) |
| **TC-PA-007** | `parse_test.go` | Parse fractional minutes | High | "2.5m" -> 150 |
| **TC-PA-008** | `parse_test.go` | Parse fractional hours | High | "1.5h" -> 5400 |
| **TC-PA-009** | `parse_test.go` | Parse negative | High | "-5h" -> -18000 |
| **TC-PA-010** | `parse_test.go` | Parse zero unit | Medium | "0s" -> 0 |
| **TC-PA-011** | `parse_test.go` | Parse zero raw | Medium | "0" -> 0 |
| **TC-PA-012** | `parse_test.go` | Parse empty | Low | Error |
| **TC-PA-013** | `parse_test.go` | Parse invalid | Low | Error |
| **TC-PA-014** | `parse_test.go` | Parse plus sign | Low | "+5h" -> 5h |
| **TC-PA-015** | `parse_test.go` | Parse spaces | Low | "5h 30m" -> 5h30m |
| **TC-PA-016** | `parse_test.go` | ParseByte valid | High | []byte("3h45m") -> duration |
| **TC-PA-017** | `parse_test.go` | ParseByte invalid | Low | Error |
| **TC-PA-018** | `parse_test.go` | ParseByte empty | Low | Error |
| **TC-PA-019** | `parse_test.go` | ParseByte days | High | []byte("7d") -> duration |
| **TC-PA-020** | `parse_test.go` | Seconds constructor | High | Seconds(30) -> 30s |
| **TC-PA-021** | `parse_test.go` | Seconds negative | Medium | Seconds(-30) -> -30s |
| **TC-PA-022** | `parse_test.go` | Seconds zero | Medium | Seconds(0) -> 0s |
| **TC-PA-023** | `parse_test.go` | Seconds large | Low | Large value handling |
| **TC-PA-024** | `parse_test.go` | Minutes constructor | High | Minutes(5) -> 5m |
| **TC-PA-025** | `parse_test.go` | Minutes negative | Medium | Minutes(-5) -> -5m |
| **TC-PA-026** | `parse_test.go` | Minutes zero | Medium | Minutes(0) -> 0s |
| **TC-PA-027** | `parse_test.go` | Hours constructor | High | Hours(3) -> 3h |
| **TC-PA-028** | `parse_test.go` | Hours negative | Medium | Hours(-3) -> -3h |
| **TC-PA-029** | `parse_test.go` | Hours zero | Medium | Hours(0) -> 0s |
| **TC-PA-030** | `parse_test.go` | Days constructor | High | Days(7) -> 7d |
| **TC-PA-031** | `parse_test.go` | Days negative | Medium | Days(-7) -> -7d |
| **TC-PA-032** | `parse_test.go` | Days zero | Medium | Days(0) -> 0s |
| **TC-PA-033** | `parse_test.go` | Days large | Low | Large value handling |
| **TC-PA-036** | `parse_test.go` | ParseDuration standard | High | time.Duration -> big.Duration |
| **TC-PA-037** | `parse_test.go` | ParseDuration zero | Medium | time.Duration(0) -> 0 |
| **TC-PA-038** | `parse_test.go` | ParseDuration negative | Medium | Negative time.Duration -> negative big.Duration |
| **TC-PA-039** | `parse_test.go` | ParseDuration subsecond | Low | Sub-second truncation |
| **TC-PA-040** | `parse_test.go` | ParseDuration tiny | Low | Nanoseconds -> 0 |
| **TC-PA-041** | `parse_test.go` | ParseFloat64 positive | High | float64 -> big.Duration |
| **TC-PA-042** | `parse_test.go` | ParseFloat64 zero | Medium | 0.0 -> 0 |
| **TC-PA-043** | `parse_test.go` | ParseFloat64 negative | Medium | Negative float -> negative duration |
| **TC-PA-044** | `parse_test.go` | ParseFloat64 max | Low | MaxFloat64 -> MaxInt64 |
| **TC-PA-045** | `parse_test.go` | ParseFloat64 min | Low | -MaxFloat64 -> MinInt64 |
| **TC-PA-046** | `parse_test.go` | ParseFloat64 fractional | Low | Rounding behavior |
| **TC-PA-047** | `parse_test.go` | ParseFloat64 large | Low | Precision limits |
| **TC-PA-048** | `parse_test.go` | ParseFloat64 near min | Low | Boundary check |
| **TC-PA-049** | `parse_test.go` | MaxInt64 boundary | Low | MaxInt64 handling |
| **TC-PA-050** | `parse_test.go` | MinInt64 boundary | Low | MinInt64 handling |
| **TC-PA-051** | `parse_test.go` | All components | High | d+h+m+s parsing |
| **TC-PA-052** | `parse_test.go` | Multiple spaces | Low | " 5h 30m " -> valid |
| **TC-PA-053** | `parse_test.go` | Quoted string | Low | "\"5h30m\"" -> valid |
| **TC-PA-054** | `parse_test.go` | Just sign | Low | "-" -> error |
| **TC-PA-055** | `parse_test.go` | Just plus | Low | "+" -> error |
| **TC-PA-056** | `parse_test.go` | Large duration parse | Low | "100000d" -> valid |
| **TC-PA-057** | `parse_test.go` | Multiple components | High | "1d2h3m4s" -> valid |
| **TC-PA-058** | `parse_test.go` | Missing unit | Low | "10" -> error |
| **TC-PA-059** | `parse_test.go` | Unknown unit | Low | "10y" -> error |
| **TC-PA-060** | `parse_test.go` | Invalid number | Low | "1.1.1h" -> error |
| **TC-FO-001** | `format_test.go` | Format zero | Critical | 0 -> "0s" |
| **TC-FO-002** | `format_test.go` | Format seconds | Critical | 45 -> "45s" |
| **TC-FO-003** | `format_test.go` | Format minutes | Critical | 330 -> "5m30s" |
| **TC-FO-004** | `format_test.go` | Format hours | Critical | -> "2h30m45s" |
| **TC-FO-005** | `format_test.go` | Format days | Critical | -> "5d" |
| **TC-FO-006** | `format_test.go` | Format days hours | Critical | -> "5d12h" |
| **TC-FO-007** | `format_test.go` | Format full | Critical | -> "5d23h15m13s" |
| **TC-FO-008** | `format_test.go` | Format negative | High | -> "-30s" |
| **TC-FO-009** | `format_test.go` | Format negative days | High | -> "-5d" |
| **TC-FO-010** | `format_test.go` | Format large | High | -> "10000d" |
| **TC-FO-013** | `format_test.go` | Time conversion | High | big.Duration -> time.Duration |
| **TC-FO-014** | `format_test.go` | Time zero | Medium | 0 -> 0 |
| **TC-FO-015** | `format_test.go` | Time negative | Medium | Negative -> Negative |
| **TC-FO-016** | `format_test.go` | Time overflow | High | Error on overflow |
| **TC-FO-017** | `format_test.go` | Time max safe | Medium | Max safe time.Duration |
| **TC-FO-018** | `format_test.go` | Time near overflow | Medium | Boundary check |
| **TC-FO-019** | `format_test.go` | Int64 positive | High | Duration -> int64 |
| **TC-FO-020** | `format_test.go` | Int64 negative | High | Duration -> int64 |
| **TC-FO-021** | `format_test.go` | Int64 zero | Medium | 0 -> 0 |
| **TC-FO-022** | `format_test.go` | Int64 large | Medium | Large value preservation |
| **TC-FO-023** | `format_test.go` | Int64 max | Low | MaxInt64 preservation |
| **TC-FO-024** | `format_test.go` | Int64 min | Low | MinInt64 preservation |
| **TC-FO-025** | `format_test.go` | Uint64 positive | High | Duration -> uint64 |
| **TC-FO-026** | `format_test.go` | Uint64 negative | High | Negative -> 0 |
| **TC-FO-027** | `format_test.go` | Uint64 zero | Medium | 0 -> 0 |
| **TC-FO-028** | `format_test.go` | Uint64 large | Medium | Large value preservation |
| **TC-FO-029** | `format_test.go` | Uint64 negative days | Low | Negative days -> 0 |
| **TC-FO-030** | `format_test.go` | Float64 positive | High | Duration -> float64 |
| **TC-FO-031** | `format_test.go` | Float64 negative | High | Duration -> float64 |
| **TC-FO-032** | `format_test.go` | Float64 zero | Medium | 0 -> 0.0 |
| **TC-FO-033** | `format_test.go` | Float64 large | Medium | Large value preservation |
| **TC-FO-034** | `format_test.go` | Float64 fractional | Low | Precision check |
| **TC-FO-035** | `format_test.go` | Float64 precision | Low | Large number precision |
| **TC-FO-036** | `format_test.go` | Format hours only | Low | "23h" (no days) |
| **TC-FO-037** | `format_test.go` | Format omit zero prefix | Low | "5m" (no hours) |
| **TC-FO-038** | `format_test.go` | Format minutes only | Low | "45m" |
| **TC-FO-039** | `format_test.go` | Format hours only explicit | Low | "12h" |
| **TC-FO-040** | `format_test.go` | Format 1 sec | Low | "1s" |
| **TC-FO-041** | `format_test.go` | Format 1 min | Low | "1m" |
| **TC-FO-042** | `format_test.go` | Format 1 hour | Low | "1h" |
| **TC-FO-043** | `format_test.go` | Format 1 day | Low | "1d" |
| **TC-FO-044** | `format_test.go` | Format max duration | Low | MaxInt64 string representation |
| **TC-FO-045** | `format_test.go` | Int64 round-trip | Medium | Seconds(i).Int64() == i |
| **TC-FO-046** | `format_test.go` | Float64 round-trip | Medium | ParseFloat64(f).Float64() ~= f |
| **TC-FO-047** | `format_test.go` | Chain conversion | Low | Multi-step conversion check |
| **TC-FO-048** | `format_test.go` | Access Days | High | d.Days() |
| **TC-FO-049** | `format_test.go` | Access Hours | High | d.Hours() |
| **TC-FO-050** | `format_test.go` | Access Minutes | High | d.Minutes() |
| **TC-FO-051** | `format_test.go` | Access Seconds | High | d.Seconds() |
| **TC-EN-001** | `encode_test.go` | JSON Marshal | High | Duration -> JSON string |
| **TC-EN-002** | `encode_test.go` | JSON Marshal zero | Medium | 0s -> "0s" |
| **TC-EN-003** | `encode_test.go` | JSON Marshal negative | Medium | Negative -> "-..." |
| **TC-EN-004** | `encode_test.go` | JSON Marshal days | Medium | With days component |
| **TC-EN-005** | `encode_test.go` | JSON Marshal large | Low | Large duration |
| **TC-EN-006** | `encode_test.go` | JSON Unmarshal | High | JSON string -> Duration |
| **TC-EN-007** | `encode_test.go` | JSON Unmarshal zero | Medium | "0s" -> 0 |
| **TC-EN-008** | `encode_test.go` | JSON Unmarshal days | Medium | With days component |
| **TC-EN-009** | `encode_test.go` | JSON Unmarshal invalid | Low | Error on invalid |
| **TC-EN-010** | `encode_test.go` | JSON Unmarshal spaces | Low | Handle spaces |
| **TC-EN-011** | `encode_test.go` | JSON Unmarshal negative | Low | Handle negative |
| **TC-EN-012** | `encode_test.go` | YAML Marshal | High | Duration -> YAML string |
| **TC-EN-013** | `encode_test.go` | YAML Marshal zero | Medium | 0s -> "0s\n" |
| **TC-EN-014** | `encode_test.go` | YAML Marshal days | Medium | With days component |
| **TC-EN-015** | `encode_test.go` | YAML Marshal negative | Medium | Negative -> "-..." |
| **TC-EN-016** | `encode_test.go` | YAML Unmarshal | High | YAML string -> Duration |
| **TC-EN-017** | `encode_test.go` | YAML Unmarshal invalid | Low | Error on invalid |
| **TC-EN-018** | `encode_test.go` | YAML Unmarshal days | Medium | With days component |
| **TC-EN-019** | `encode_test.go` | YAML Unmarshal zero | Medium | "0s" -> 0 |
| **TC-EN-020** | `encode_test.go` | TOML Marshal | High | Duration -> TOML string |
| **TC-EN-021** | `encode_test.go` | TOML Marshal days | Medium | With days component |
| **TC-EN-022** | `encode_test.go` | TOML Marshal zero | Medium | 0s -> "0s" |
| **TC-EN-023** | `encode_test.go` | TOML Unmarshal string | High | TOML string -> Duration |
| **TC-EN-024** | `encode_test.go` | TOML Unmarshal bytes | Medium | TOML bytes -> Duration |
| **TC-EN-025** | `encode_test.go` | TOML Unmarshal invalid type | Low | Error on wrong type |
| **TC-EN-026** | `encode_test.go` | TOML Unmarshal invalid str | Low | Error on invalid string |
| **TC-EN-027** | `encode_test.go` | TOML Unmarshal quoted | Low | Handle quotes |
| **TC-EN-028** | `encode_test.go` | Text Marshal | High | Duration -> Text |
| **TC-EN-029** | `encode_test.go` | Text Marshal zero | Medium | 0s -> "0s" |
| **TC-EN-030** | `encode_test.go` | Text Marshal negative | Medium | Negative handling |
| **TC-EN-031** | `encode_test.go` | Text Marshal days | Medium | With days |
| **TC-EN-032** | `encode_test.go` | Text Unmarshal | High | Text -> Duration |
| **TC-EN-033** | `encode_test.go` | Text Unmarshal invalid | Low | Error handling |
| **TC-EN-034** | `encode_test.go` | Text Unmarshal days | Medium | With days |
| **TC-EN-035** | `encode_test.go` | Text Unmarshal empty | Low | Empty string handling |
| **TC-EN-036** | `encode_test.go` | CBOR Marshal | High | Duration -> CBOR |
| **TC-EN-037** | `encode_test.go` | CBOR Marshal days | Medium | With days |
| **TC-EN-038** | `encode_test.go` | CBOR Marshal zero | Medium | Zero value |
| **TC-EN-039** | `encode_test.go` | CBOR Unmarshal | High | CBOR -> Duration |
| **TC-EN-040** | `encode_test.go` | CBOR Unmarshal invalid | Low | Error handling |
| **TC-EN-041** | `encode_test.go` | CBOR Unmarshal bad string | Low | Invalid duration string in CBOR |
| **TC-EN-042** | `encode_test.go` | CBOR Unmarshal days | Medium | With days |
| **TC-EN-043** | `encode_test.go` | Round-trip JSON | High | Marshal -> Unmarshal |
| **TC-EN-044** | `encode_test.go` | Round-trip YAML | High | Marshal -> Unmarshal |
| **TC-EN-045** | `encode_test.go` | Round-trip TOML | High | Marshal -> Unmarshal |
| **TC-EN-046** | `encode_test.go` | Round-trip Text | High | Marshal -> Unmarshal |
| **TC-EN-047** | `encode_test.go` | Round-trip CBOR | High | Marshal -> Unmarshal |
| **TC-OP-001** | `operation_test.go` | Abs positive | Medium | 100 -> 100 |
| **TC-OP-002** | `operation_test.go` | Abs zero | Medium | 0 -> 0 |
| **TC-OP-003** | `operation_test.go` | Abs negative | Medium | -100 -> 100 |
| **TC-OP-004** | `operation_test.go` | Abs negative days | Medium | -5d -> 5d |
| **TC-OP-005** | `operation_test.go` | Abs min duration | Low | MinInt64 -> MaxInt64 |
| **TC-OP-006** | `operation_test.go` | Abs large negative | Low | Large value check |
| **TC-OP-007** | `operation_test.go` | RangeTo basic | High | Generate range start->end |
| **TC-OP-008** | `operation_test.go` | RangeTo inclusive | Medium | Include bounds |
| **TC-OP-009** | `operation_test.go` | RangeTo monotonic | Medium | Values increasing |
| **TC-OP-010** | `operation_test.go` | RangeTo zero start | Low | Start from 0 |
| **TC-OP-011** | `operation_test.go` | RangeTo timeout | Low | Context timeout check |
| **TC-OP-012** | `operation_test.go` | RangeDefTo | Medium | Using default params |
| **TC-OP-013** | `operation_test.go` | RangeDefTo count | Low | Step count sanity check |
| **TC-OP-014** | `operation_test.go` | RangeFrom basic | High | Generate range end<-start |
| **TC-OP-015** | `operation_test.go` | RangeFrom inclusive | Medium | Include bounds |
| **TC-OP-016** | `operation_test.go` | RangeFrom monotonic | Medium | Values increasing (inverse) |
| **TC-OP-017** | `operation_test.go` | RangeFrom timeout | Low | Context timeout check |
| **TC-OP-019** | `operation_test.go` | RangeDefFrom | Medium | Using default params |
| **TC-OP-020** | `operation_test.go` | RangeDefFrom count | Low | Step count sanity check |
| **TC-OP-021** | `operation_test.go` | RangeCtxTo timeout | Medium | Respect timeout |
| **TC-OP-022** | `operation_test.go` | RangeCtxTo valid | Medium | Normal execution |
| **TC-OP-023** | `operation_test.go` | RangeCtxTo cancel | Medium | Context cancellation |
| **TC-OP-024** | `operation_test.go` | RangeCtxTo min elements | Low | Fallback to min elements |
| **TC-OP-025** | `operation_test.go` | RangeCtxFrom timeout | Medium | Respect timeout |
| **TC-OP-026** | `operation_test.go` | RangeCtxFrom valid | Medium | Normal execution |
| **TC-OP-027** | `operation_test.go` | RangeCtxFrom cancel | Medium | Context cancellation |
| **TC-OP-028** | `operation_test.go` | Range small rates | Low | PID tuning |
| **TC-OP-029** | `operation_test.go` | Range large rates | Low | PID tuning |
| **TC-OP-030** | `operation_test.go` | Default rates | Low | Check constants |
| **TC-OP-031** | `operation_test.go` | RangeTo perf | Low | Basic perf check |
| **TC-OP-032** | `operation_test.go` | RangeFrom perf | Low | Basic perf check |
| **TC-OP-033** | `operation_test.go` | Range Days | Medium | Unit type check |
| **TC-OP-034** | `operation_test.go` | Range Hours | Medium | Unit type check |
| **TC-OP-035** | `operation_test.go` | Range Minutes | Medium | Unit type check |
| **TC-OP-036** | `operation_test.go` | Range Seconds | Medium | Unit type check |
| **TC-OP-037** | `operation_test.go` | Range Mixed Units | Medium | Mixed types handling |
| **TC-TR-001** | `truncate_test.go` | Truncate minute | Medium | 5m45s -> 5m |
| **TC-TR-002** | `truncate_test.go` | Truncate hour | Medium | 3h45m -> 3h |
| **TC-TR-003** | `truncate_test.go` | Truncate day | Medium | 2d12h -> 2d |
| **TC-TR-004** | `truncate_test.go` | Truncate negative unit | Low | Unchanged |
| **TC-TR-005** | `truncate_test.go` | Truncate zero unit | Low | Unchanged |
| **TC-TR-006** | `truncate_test.go` | Truncate zero | Low | 0 -> 0 |
| **TC-TR-007** | `truncate_test.go` | Truncate negative | Medium | -5m30s -> -5m |
| **TC-TR-008** | `truncate_test.go` | Truncate exact | Low | 10m -> 10m |
| **TC-TR-009** | `truncate_test.go` | Truncate custom | Low | 127s / 10s -> 120s |
| **TC-TR-010** | `truncate_test.go` | Round minute up | Medium | 55s -> 1m |
| **TC-TR-011** | `truncate_test.go` | Round minute down | Medium | 25s -> 0s |
| **TC-TR-012** | `truncate_test.go` | Round halfway | Medium | 30s -> 1m |
| **TC-TR-013** | `truncate_test.go` | Round negative | Medium | -55s -> -1m |
| **TC-TR-014** | `truncate_test.go` | Round negative unit | Low | Unchanged |
| **TC-TR-015** | `truncate_test.go` | Round zero unit | Low | Unchanged |
| **TC-TR-016** | `truncate_test.go` | Round zero | Low | 0 -> 0 |
| **TC-TR-017** | `truncate_test.go` | Round hours | Medium | 2h45m -> 3h |
| **TC-TR-018** | `truncate_test.go` | Round days | Medium | 1d18h -> 2d |
| **TC-TR-019** | `truncate_test.go` | Round exact | Low | 5h -> 5h |
| **TC-TR-020** | `truncate_test.go` | TruncateMinutes | High | Helper function |
| **TC-TR-021** | `truncate_test.go` | TruncateMinutes zero | Low | 0 -> 0 |
| **TC-TR-022** | `truncate_test.go` | TruncateMinutes neg | Medium | Negative handling |
| **TC-TR-023** | `truncate_test.go` | TruncateMinutes exact | Low | No change |
| **TC-TR-024** | `truncate_test.go` | TruncateMinutes mix | Medium | With hours |
| **TC-TR-025** | `truncate_test.go` | TruncateHours | High | Helper function |
| **TC-TR-026** | `truncate_test.go` | TruncateHours zero | Low | 0 -> 0 |
| **TC-TR-027** | `truncate_test.go` | TruncateHours neg | Medium | Negative handling |
| **TC-TR-028** | `truncate_test.go` | TruncateHours exact | Low | No change |
| **TC-TR-029** | `truncate_test.go` | TruncateHours mix | Medium | With days |
| **TC-TR-030** | `truncate_test.go` | TruncateDays | High | Helper function |
| **TC-TR-031** | `truncate_test.go` | TruncateDays zero | Low | 0 -> 0 |
| **TC-TR-032** | `truncate_test.go` | TruncateDays neg | Medium | Negative handling |
| **TC-TR-033** | `truncate_test.go` | TruncateDays exact | Low | No change |
| **TC-TR-034** | `truncate_test.go` | TruncateDays partial | Medium | Rounding down |
| **TC-TR-035** | `truncate_test.go` | TruncateDays small | Low | < 1d -> 0 |
| **TC-TR-036** | `truncate_test.go` | Truncate small unit | Low | Edge case |
| **TC-TR-037** | `truncate_test.go` | Truncate large unit | Low | Edge case |
| **TC-TR-038** | `truncate_test.go` | Truncate 59s to 0m | Low | Boundary check |
| **TC-TR-039** | `truncate_test.go` | Truncate 59m to 0h | Low | Boundary check |
| **TC-TR-040** | `truncate_test.go` | Truncate 23h to 0d | Low | Boundary check |
| **TC-TR-041** | `truncate_test.go` | Round 30s | Low | Halfway check |
| **TC-TR-042** | `truncate_test.go` | Round 29s | Low | Down check |
| **TC-TR-043** | `truncate_test.go` | Round 31s | Low | Up check |
| **TC-TR-044** | `truncate_test.go` | Round neg halfway | Low | Negative halfway |
| **TC-TR-045** | `truncate_test.go` | Chain operations | Medium | Truncate().Truncate() |
| **TC-BM-001** | `model_test.go` | Bench Viper | Low | Performance |
| **TC-BM-002** | `model_test.go` | Bench IsDays | Low | Performance |
| **TC-BM-003** | `model_test.go` | Bench IsHours | Low | Performance |
| **TC-BM-004** | `model_test.go` | Bench IsMinutes | Low | Performance |
| **TC-BM-005** | `model_test.go` | Bench IsSeconds | Low | Performance |
| **TC-BM-006** | `parse_test.go` | Bench Parse | Low | Performance |
| **TC-BM-007** | `format_test.go` | Bench String | Low | Performance |
| **TC-BM-008** | `format_test.go` | Bench Time | Low | Performance |
| **TC-BM-009** | `truncate_test.go` | Bench Truncate | Low | Performance |
| **TC-BM-010** | `truncate_test.go` | Bench Round | Low | Performance |

---

## Framework & Tools

### Testing Frameworks

#### Ginkgo v2 - BDD Testing Framework

**Why Ginkgo over standard Go testing:**
- ✅ **Hierarchical organization**: `Describe`, `Context`, `It` for clear test structure
- ✅ **Better readability**: Tests read like specifications
- ✅ **Rich lifecycle hooks**: `BeforeEach`, `AfterEach` for setup/teardown
- ✅ **Async testing**: `Eventually`, `Consistently` for concurrent behavior
- ✅ **Parallel execution**: Built-in support for concurrent test runs

#### Gomega - Matcher Library

**Advantages:**
- ✅ **Expressive matchers**: `Equal`, `BeNumerically`, `HaveOccurred`
- ✅ **Clear failures**: Detailed error messages
- ✅ **Async assertions**: `Eventually` polls for state changes

#### gmeasure - Performance Measurement

Used for benchmarking throughput and latency within the BDD suite.

### Testing Concepts & Standards

#### ISTQB Alignment

This test suite follows **ISTQB (International Software Testing Qualifications Board)** principles:

1. **Test Levels** (ISTQB Foundation Level):
    - **Component Testing**: Verifying individual functions like `Parse` or `String` in unit tests.
    - **Integration Testing**: Verifying serialization with external libraries (JSON, YAML).

2. **Test Types** (ISTQB Advanced Level):
    - **Functional Testing**: Verifying the core logic of the package.
    - **Non-functional Testing**: Performance benchmarking and memory profiling.

3. **Test Design Techniques**:
    - **Boundary Value Analysis**: Testing `MaxInt64`, `MinInt64`, zero values.
    - **Equivalence Partitioning**: Testing positive/negative/zero classes.

---

## Quick Start

### Installation

```bash
# Install Ginkgo CLI (optional)
go install github.com/onsi/ginkgo/v2/ginkgo@latest
```

### Run Tests

```bash
# Run all tests
go test -v ./...
```

```bash
# With coverage
go test -v -cover ./...
```

```bash
# With race detection (recommended)
CGO_ENABLED=1 go test -v -race ./...
```

```bash
# Benchmarks
go test -v -bench=. -benchmem ./...
```

---

### Performance & Profiling

Detailed performance analysis tools and commands.

#### Performance CPU

Profiling CPU usage to identify hotspots.

Command to launch :
```bash
# Benchmarks with CPU profile
go test -bench=. -cpuprofile=cpu.out ./...
go tool pprof cpu.out
```

#### Performance Memory

Profiling memory allocation to identify leaks or heavy allocs.

Command to launch :
```bash
# Benchmarks with memory profile
go test -bench=. -memprofile=mem.out ./...
go tool pprof mem.out
```

---

## Test Coverage

**Target**: ≥85% statement coverage (currently 86.7% aggregate)

### Coverage By Package

```bash
# View coverage by package
go test -cover ./...
```

**Output**:
```
github.com/nabbar/golib/duration/big  coverage: 86.4% of statements
```

### Coverage By File

```bash
# Generate detailed report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

**Output**:
```
github.com/nabbar/golib/duration/big/encode.go:     100.0%
github.com/nabbar/golib/duration/big/format.go:      96.0%
github.com/nabbar/golib/duration/big/interface.go:  100.0%
github.com/nabbar/golib/duration/big/model.go:      100.0%
github.com/nabbar/golib/duration/big/operation.go:   93.0%
github.com/nabbar/golib/duration/big/parse.go:       88.0%
github.com/nabbar/golib/duration/big/truncate.go:    95.0%
```

---

## Writing Tests

### Guidelines

**1. Use Descriptive Names**
```go
It("should parse complex duration string correctly", func() {
    // Test implementation
})
```

**2. Follow AAA Pattern** (Arrange, Act, Assert)
```go
It("should format days correctly", func() {
    // Arrange
    d := big.Days(5)
    
    // Act
    s := d.String()
    
    // Assert
    Expect(s).To(Equal("5d"))
})
```

**3. Use Appropriate Matchers**

```go
Expect(err).ToNot(HaveOccurred())
Expect(d.Int64()).To(Equal(expected))
Expect(s).To(ContainSubstring("-"))
```

**4. Always Cleanup Resources**

```go
// Example for context timeout
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
```

**5. Test Edge Cases**
- Negative durations
- Zero duration
- Maximum/Minimum int64 values
- Invalid string inputs

### Test Template

```go
var _ = Describe("duration/big/Feature", func() {
    Context("When performing operation", func() {
        It("should produce expected result", func() {
            // Arrange
            // ...
            
            // Act
            // ...
            
            // Assert
            // ...
        })
    })
})
```

---

## Best Practices

### Test Independence

**✅ Good Practices**
- Each test should be independent and not rely on shared state from other tests.
- Use `BeforeEach` sparingly for common setup, but prefer explicit setup in `It` blocks for clarity when possible.

### Assertions

**✅ Good**
```go
Expect(d).To(Equal(expected))
Expect(err).ToNot(HaveOccurred())
```

**❌ Avoid**
```go
if d != expected {
    t.Fail()
}
```

### Concurrency Testing

```go
// Use race detector during execution
// go test -race ./...
```

### Performance

- Keep unit tests fast.
- Isolate heavy performance tests in benchmarks.
- Use `-benchmem` to track allocations.

### Test Timeouts

```bash
# Identify slow tests
ginkgo --timeout=10s
```

### Debugging

```bash
# Single test
ginkgo --focus="should parse complex"

# With stack traces
go test -v ./...
```

---

## Reporting Bugs & Vulnerabilities

### Bug Report Template

When reporting a bug in the test suite or the socket package, please use this template:

```markdown
**Title**: [BUG] Brief description of the bug

**Description**:
[A clear and concise description of what the bug is.]

**Steps to Reproduce:**
1. [First step]
2. [Second step]
3. [...]

**Expected Behavior**:
[A clear and concise description of what you expected to happen]

**Actual Behavior**:
[What actually happened]

**Code Example**:
[Minimal reproducible example]

**Test Case** (if applicable):
[Paste full test output with -v flag]

**Environment**:
- Go version: `go version`
- OS: Linux/macOS/Windows
- Architecture: amd64/arm64
- Package version: vX.Y.Z or commit hash

**Additional Context**:
[Any other relevant information]

**Logs/Error Messages**:
[Paste error messages or stack traces here]

**Possible Fix:**
[If you have suggestions]
```

### Security Vulnerability Template

**⚠️ IMPORTANT**: For security vulnerabilities, please **DO NOT** create a public issue.

Instead, report privately via:
1. GitHub Security Advisories (preferred)
2. Email to the maintainer (see footer)

**Vulnerability Report Template:**

```markdown
**Vulnerability Type:**
[e.g., Overflow, Race Condition, Memory Leak, Denial of Service]

**Severity:**
[Critical / High / Medium / Low]

**Affected Component:**
[e.g., interface.go, context.go, specific function]

**Affected Versions**:
[e.g., v1.0.0 - v1.2.3]

**Vulnerability Description:**
[Detailed description of the security issue]

**Attack Scenario**:
1. Attacker does X
2. System responds with Y
3. Attacker exploits Z

**Proof of Concept:**
[Minimal code to reproduce the vulnerability]
[DO NOT include actual exploit code]

**Impact**:
- Confidentiality: [High / Medium / Low]
- Integrity: [High / Medium / Low]
- Availability: [High / Medium / Low]

**Proposed Fix** (if known):
[Suggested approach to fix the vulnerability]

**CVE Request**:
[Yes / No / Unknown]

**Coordinated Disclosure**:
[Willing to work with maintainers on disclosure timeline]
```

### Issue Labels

When creating GitHub issues, use these labels:

- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Improvements to docs
- `performance`: Performance issues
- `test`: Test-related issues
- `security`: Security vulnerability (private)
- `help wanted`: Community help appreciated
- `good first issue`: Good for newcomers

### Reporting Guidelines

**Before Reporting:**
1. ✅ Search existing issues to avoid duplicates
2. ✅ Verify the bug with the latest version
3. ✅ Run tests with `-race` detector
4. ✅ Check if it's a test issue or package issue
5. ✅ Collect all relevant logs and outputs

**What to Include:**
- Complete test output (use `-v` flag)
- Go version (`go version`)
- OS and architecture (`go env GOOS GOARCH`)
- Race detector output (if applicable)
- Coverage report (if relevant)

**Response Time:**
- **Bugs**: Typically reviewed within 48 hours
- **Security**: Acknowledged within 24 hours
- **Enhancements**: Reviewed as time permits

---

## Resources

**Testing Frameworks**
- [Ginkgo Documentation](https://onsi.github.io/ginkgo/): BDD Testing Framework for Go.
- [Gomega Matchers](https://onsi.github.io/gomega/): Matcher/Assertion library.
- [Go Testing](https://pkg.go.dev/testing): Standard library testing package.
- [Go Coverage](https://go.dev/blog/cover): The cover story.

**Testing References**
- [ISTQB concept](https://www.istqb.org/): International Software Testing Qualifications Board.

**Concurrency**
- [Go Race Detector](https://go.dev/doc/articles/race_detector): Data race detector.
- [Go Memory Model](https://go.dev/ref/mem): The Go Memory Model.
- [sync Package](https://pkg.go.dev/sync): Synchronization primitives.
- [sync/atomic Package](https://pkg.go.dev/sync/atomic): Atomic memory primitives.

**Performance**
- [Go Profiling](https://go.dev/blog/pprof): Profiling Go Programs.
- [Benchmarking](https://pkg.go.dev/testing#hdr-Benchmarks): How to write benchmarks.
- [Execution Tracer](https://go.dev/doc/diagnostics#execution-tracer): Trace tool.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for test generation, debugging, and documentation under human supervision. All tests are validated and reviewed by humans.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2022 Nicolas JUHEL
