# OA_Context Documentation

## Overview
The `OA_Context` structure is the central data orchestration mechanism within the `oa` remastering engine. It facilitates a "Parameter Inheritance" design pattern, allowing the engine to seamlessly handle both global configurations and specific task overrides during the execution of a JSON-driven workflow.

## Structure Definition
Defined in `include/oa.h`, the struct utilizes the `cJSON` library to manage memory-efficient pointers to the parsed workflow plan:

```c
typedef struct {
    cJSON *root;  // Global configuration (the entire JSON plan)
    cJSON *task;  // Local configuration (the specific task/command being executed)
} OA_Context;
```

## The "Cascade Lookup" Pattern
The dual-pointer architecture of `OA_Context` solves the problem of parameter scoping. When an action module (e.g., `action_squash` or `action_initrd`) needs a parameter like `pathLiveFs`, it employs a cascade lookup strategy:

1.  **Local Scope (`ctx->task`)**: The action first checks if the parameter is explicitly defined inside the current task block. This allows for granular overrides for specific steps in the pipeline (e.g., passing unique arguments to `action_run`).
2.  **Global Scope (`ctx->root`)**: If the parameter is not found locally, the action automatically falls back to the root JSON object. This prevents redundancy in the `plan.json`, as universal variables like `pathLiveFs` only need to be declared once at the top level.

## Lifecycle and Routing
The context is instantiated inside `src/main.c` within the `execute_verb` function, acting as a traffic controller for the engine.

* The engine parses the incoming JSON file and iterates through the `plan` array.
* For each task in the array, a new `OA_Context` is created on the stack, binding the `root` JSON and the specific `task` JSON.
* This context is then passed by reference (`&ctx`) to the designated action function based on the `command` key.

## Standard Implementation Example
Action modules across `src/actions/` uniformly apply this context to extract data safely. Here is the standard implementation pattern:

```c
int action_example(OA_Context *ctx) {
    // 1. Attempt to fetch parameter from the local task 
    cJSON *pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->task, "pathLiveFs");
    
    // 2. Fallback to the global root if not found locally
    if (!pathLiveFs) {
        pathLiveFs = cJSON_GetObjectItemCaseSensitive(ctx->root, "pathLiveFs");
    }

    // 3. Validate the extracted data before proceeding 
    if (!cJSON_IsString(pathLiveFs)) {
        return 1;
    }
    
    // Proceed with native kernel syscalls and action logic...
    return 0;
}
```