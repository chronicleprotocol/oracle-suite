variables {
  word_one = "hello"
  word_two = "world"
  greeting = "${var.word_one} ${var.word_two}"

  map_values = {
    integer_a  = 1
    integer_b  = 2
    nested_map = {
      integer_d = 3
      integer_e = 4
    }
    nested_list = [5, 6]
  }

  list_values = [1, 2, [3, 4], { "key_x" = 5, "key_y" = 6 }]

  simple_ref1 = var.map_values.integer_a
  simple_ref2 = var.map_values.integer_b
  simple_ref3 = var.list_values[0]
  simple_ref4 = var.list_values[1]
  simple_ref5 = "hello ${var.map_values.integer_a}"
  simple_ref6 = [var.map_values.integer_a, var.map_values.integer_b]
  simple_ref7 = {
    ref_a = var.map_values.integer_a
    ref_b = var.map_values.integer_b
  }

  complex_ref1 = {
    nested_map_ref = {
      ref_x = var.map_values.nested_map.integer_d
      ref_y = var.map_values.nested_map.integer_e
    }
    nested_list_ref = var.map_values.nested_list
  }

  complex_ref2 = {
    nested_list_map_ref  = var.list_values[3]
    nested_list_item_ref = var.list_values[2][1]
  }

  complex_ref3 = {
    string_interpolation_ref = "Value: ${var.map_values.nested_map.integer_d}, List Item: ${var.list_values[2][0]}"
    map_interpolation_ref    = {
      ref_a = "X: ${var.list_values[3].key_x}, Y: ${var.list_values[3].key_y}"
    }
  }

  complex_ref4 = {
    integer_a = 1
    integer_b = var.complex_ref4.integer_a
  }

  empty_list = []
  empty_map  = {}
}
