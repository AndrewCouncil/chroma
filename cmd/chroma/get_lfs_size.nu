#!/usr/bin/env nu

let lfs_file_data = (
  git lfs ls-files --size |
    detect columns -n |
    reject column1 |
    rename hash file size |
    update size {|row| $row.size | str substring 1..-2 | into filesize} |
    sort-by size
)

print $lfs_file_data

let total_size = ($lfs_file_data | get size | math sum)
print $"Total size: ($total_size)"
