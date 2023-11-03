variables {
  ghost_pairs = explode(var.item_separator, env("CFG_SYMBOLS", env("CFG_GHOST_PAIRS", "")))
}

ghost {
  ethereum_key = "default"
  interval     = tonumber(env("CFG_GHOST_INTERVAL", "60"))
  data_models  = distinct(concat([
    for v in var.contracts : v.wat
    # Limit the list only to a specific environment but take all chains
    if v.env == var.environment
    # Only Scribe compatible contracts
    && try(v.is_scribe, false)
    # If CFG_GHOST_PAIRS is set to a list of asset symbols, only for those assets will the signatures be created
    && try(length(var.ghost_pairs) == 0 || contains(var.ghost_pairs, v.wat), false)
  ], [
    for v in var.contracts : replace(v.wat, "/", "")
    # Limit the list only to a specific environment but take all chains
    if v.env == var.environment
    # Only Scribe compatible contracts
    && try(v.is_median, false)
    # If CFG_GHOST_PAIRS is set to a list of asset symbols, only for those assets will the signatures be created
    && try(length(var.ghost_pairs) == 0 || contains(var.ghost_pairs, v.wat), false)
  ], [
    for v in var.models : v
    if try(length(var.ghost_pairs) == 0 || contains(var.ghost_pairs, v), false)
  ]))
}
