# pocket-obsidian-migrator
A command line tool for converting Pocket export files to Obsidian Markdown Files

## Usage

To use the CLI you can run the following command:

```bash
./pocket-obsidian-migrator import -f /path/to/pocket_export.csv -o /path/to/output_directory
```

It will create a subdirectory called `clippings` in the output directory and write the converted Markdown files there.
A `failed.csv` file will also be created in the output directory containing any entries that could not be converted and 
the reason for the failure.

To clear down the output directory before running the import you can use the clear command:

```bash
./pocket-obsidian-migrator clear -o /path/to/output_directory
```
To see verbose output during the import process, you can use the `-v` flag:

```bash
./pocket-obsidian-migrator import -f /path/to/pocket_export.csv -o /path/to/output_directory -v
```
