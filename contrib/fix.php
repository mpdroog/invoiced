<?php
/**
 * Tool to remove weird control character ^N 0x0e
 */
$dir = "./acct/path/to/broken/files";
$rii = new RecursiveIteratorIterator(new RecursiveDirectoryIterator($dir));
foreach ($rii as $file) {
    if (in_array($file->getFilename(), [".", "..", "findreplace.php", ".git"])) continue; // Ignored files (in curdir)
    if (strpos($file->getPath(), ".git") !== false) continue; // Ignore everything in the .git dir
    if (strpos($file->getFilename(), ".toml") === false) continue; // Only parse .php files

    echo $file . "\n";

    // remove control chars
    $str = file_get_contents($file);
    $clean = preg_replace('/[^\PC\s]/u', '', $str);
    file_put_contents($file, $clean);
}

