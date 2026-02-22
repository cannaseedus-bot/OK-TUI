#include <stdio.h>
#include "agl_runtime.h"

void agl_flux_phase_enter(const char* phase)
{
    printf("[FLUX PHASE] %s\n", phase);
}

void agl_flux_barrier(const char* barrier)
{
    printf("[FLUX BARRIER] %s\n", barrier);
}

void agl_glyph_exec(const char* glyph)
{
    printf("[GLYPH EXEC] %s\n", glyph);
}
