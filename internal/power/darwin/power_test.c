#define TESTING
#include "power.h"
#include <assert.h>
#include <math.h>
#include <stdio.h>
#include <string.h>

// Mock SMC command structure for testing
static smc_cmd_t create_test_cmd(uint32_t data_type, uint8_t byte1,
                                 uint8_t byte2) {
  smc_cmd_t cmd = {0};
  cmd.keyInfo = data_type;
  cmd.data[0] = byte1;
  cmd.data[1] = byte2;
  return cmd;
}

// Test helper to compare floats with tolerance
static bool float_equals(float a, float b, float epsilon) {
  return fabsf(a - b) < epsilon;
}

// Test cases for each SMC float type
void test_fp1f_decoding(void) {
  // Test value: 1.5 (1 + 0.5)
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FP1F, 0x01, 0x40);
  float result = decode_smc_float(&cmd);
  printf("FP1F test: expected 1.5, got %f\n", result);
  assert(float_equals(result, 1.5f, 0.0001f));
  printf("FP1F test passed\n");
}

void test_fp4c_decoding(void) {
  // Test value: 4.25
  // For FP4C: value = fraction * 2^exponent / 2^12
  // 4.25 = 17408 / 4096 = 0x4400 / 0x1000
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FP4C, 0x44, 0x00);
  float result = decode_smc_float(&cmd);
  printf("FP4C test: expected 4.25, got %f\n", result);
  assert(float_equals(result, 4.25f, 0.0001f));
  printf("FP4C test passed\n");
}

void test_fp88_decoding(void) {
  // Test value: 1.5 (1 + 128/256)
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FP88, 0x01, 0x80);
  float result = decode_smc_float(&cmd);
  printf("FP88 test: expected 1.5, got %f\n", result);
  assert(float_equals(result, 1.5f, 0.0001f));
  printf("FP88 test passed\n");
}

void test_fp5b_decoding(void) {
  // Test value: 3.75 (3 + 0.75)
  // For FP5B: value = fraction * 2^exponent / 2^11
  // 3.75 = 7680 / 2048 = 0x1E00 / 0x800
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FP5B, 0x1E, 0x00);
  float result = decode_smc_float(&cmd);
  printf("FP5B test: expected 3.75, got %f\n", result);
  assert(float_equals(result, 3.75f, 0.0001f));
  printf("FP5B test passed\n");
}

void test_fp6a_decoding(void) {
  // Test value: 2.625 (2 + 0.625)
  // For FP6A: value = fraction * 2^exponent / 2^10
  // 2.625 = 2688 / 1024 = 0x0A80 / 0x400
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FP6A, 0x0A, 0x80);
  float result = decode_smc_float(&cmd);
  printf("FP6A test: expected 2.625, got %f\n", result);
  assert(float_equals(result, 2.625f, 0.0001f));
  printf("FP6A test passed\n");
}

void test_fp79_decoding(void) {
  // Test value: 5.125 (5 + 0.125)
  // For FP79: value = fraction * 2^exponent / 2^9
  // 5.125 = 2624 / 512 = 0x0A40 / 0x200
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FP79, 0x0A, 0x40);
  float result = decode_smc_float(&cmd);
  printf("FP79 test: expected 5.125, got %f\n", result);
  assert(float_equals(result, 5.125f, 0.0001f));
  printf("FP79 test passed\n");
}

void test_fpa6_decoding(void) {
  // Test value: 10.25 (10 + 0.25)
  // For FPA6: value = integer + fraction/64
  // 10.25 = 10 + 16/64 = 0x0290
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FPA6, 0x02, 0x90);
  float result = decode_smc_float(&cmd);
  printf("FPA6 test: expected 10.25, got %f\n", result);
  assert(float_equals(result, 10.25f, 0.0001f));
  printf("FPA6 test passed\n");
}

void test_fpc4_decoding(void) {
  // Test value: 15.125 (15 + 0.125)
  // For FPC4: value = integer + fraction/16
  // 15.125 = 15 + 2/16 = 0x0F02
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FPC4, 0x0F, 0x02);
  float result = decode_smc_float(&cmd);
  printf("FPC4 test: expected 15.125, got %f\n", result);
  assert(float_equals(result, 15.125f, 0.0001f));
  printf("FPC4 test passed\n");
}

void test_fpe2_decoding(void) {
  // Test value: 20.25 (20 + 0.25)
  // For FPE2: value = integer + fraction/4
  // 20.25 = 20 + 1/4 = 0x1401
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FPE2, 0x14, 0x01);
  float result = decode_smc_float(&cmd);
  printf("FPE2 test: expected 20.25, got %f\n", result);
  assert(float_equals(result, 20.25f, 0.0001f));
  printf("FPE2 test passed\n");
}

void test_edge_cases(void) {
  float result;

  // Test NULL command
  result = decode_smc_float(NULL);
  printf("NULL test: expected 0.0, got %f\n", result);
  assert(float_equals(result, 0.0f, 0.0001f));

  // Test zero values
  smc_cmd_t cmd = create_test_cmd(SMC_TYPE_FP88, 0x00, 0x00);
  result = decode_smc_float(&cmd);
  printf("Zero test: expected 0.0, got %f\n", result);
  assert(float_equals(result, 0.0f, 0.0001f));

  // Test invalid type
  cmd = create_test_cmd(0x12345678, 0x01, 0x80);
  result = decode_smc_float(&cmd);
  printf("Invalid type test: expected 0.0, got %f\n", result);
  assert(float_equals(result, 0.0f, 0.0001f));

  printf("Edge cases tests passed\n");
}

#ifdef RUN_C_TESTS
int main(void) {
  printf("Running SMC float decoding tests...\n");

  test_fp1f_decoding();
  test_fp4c_decoding();
  test_fp88_decoding();
  test_fp5b_decoding();
  test_fp6a_decoding();
  test_fp79_decoding();
  test_fpa6_decoding();
  test_fpc4_decoding();
  test_fpe2_decoding();
  test_edge_cases();

  printf("All tests passed!\n");
  return 0;
}
#endif