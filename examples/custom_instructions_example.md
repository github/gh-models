# Custom Instructions Example

This example demonstrates how to use custom instructions with the `gh models generate` command to customize the behavior of each generation phase.

## Usage

The generate command now supports custom system instructions for each phase:

```bash
# Customize intent generation
gh models generate --instruction-intent "Focus on the business value and user goals" prompt.yml

# Customize input specification generation  
gh models generate --instruction-inputspec "Include data types, validation rules, and example values" prompt.yml

# Customize output rules generation
gh models generate --instruction-outputrules "Prioritize security and performance requirements" prompt.yml

# Customize inverse output rules generation
gh models generate --instruction-inverseoutputrules "Focus on common failure modes and edge cases" prompt.yml

# Customize tests generation
gh models generate --instruction-tests "Generate comprehensive edge cases and security-focused test scenarios" prompt.yml

# Use multiple custom instructions together
gh models generate \
  --instruction-intent "Focus on the business value and user goals" \
  --instruction-inputspec "Include data types, validation rules, and example values" \
  --instruction-outputrules "Prioritize security and performance requirements" \
  --instruction-inverseoutputrules "Focus on common failure modes and edge cases" \
  --instruction-tests "Generate comprehensive edge cases and security-focused test scenarios" \
  prompt.yml
```

## What Happens

When you provide custom instructions, they are added as additional system prompts before the default instructions for each phase:

1. **Intent Phase**: Your custom intent instruction is added before the default "Analyze the following prompt and describe its intent in 2-3 sentences."

2. **Input Specification Phase**: Your custom inputspec instruction is added before the default "Analyze the following prompt and generate a specification for its inputs."

3. **Output Rules Phase**: Your custom outputrules instruction is added before the default "Analyze the following prompt and generate a list of output rules."

4. **Inverse Output Rules Phase**: Your custom inverseoutputrules instruction is added before the default "Based on the following <output_rules>, generate inverse rules that describe what would make an INVALID output."

5. **Tests Generation Phase**: Your custom tests instruction is added before the default tests generation prompt.

## Example Custom Instructions

Here are some examples of useful custom instructions for different types of prompts:

### For API Documentation Prompts
```bash
--instruction-intent "Focus on developer experience and API usability"
--instruction-inputspec "Include parameter types, required/optional status, and authentication requirements"
--instruction-outputrules "Ensure responses follow REST API conventions and include proper HTTP status codes"
```

### For Creative Writing Prompts
```bash
--instruction-intent "Emphasize creativity, originality, and narrative flow"
--instruction-inputspec "Specify genre, tone, character requirements, and length constraints"
--instruction-outputrules "Focus on story structure, character development, and engaging prose"
```

### For Code Generation Prompts
```bash
--instruction-intent "Prioritize code quality, maintainability, and best practices"
--instruction-inputspec "Include programming language, framework versions, and dependency requirements"
--instruction-outputrules "Ensure code follows language conventions, includes error handling, and has proper documentation"
```
