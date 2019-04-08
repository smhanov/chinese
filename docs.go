/*
Package chinese provides utilities for dealing with Chinese text, including
text segmentation.

Chinese text is commonly written without any spaces between the words.
This package uses the viterbi algorithm and word frequency information
to find the best placement of spaces in the sentences.

It is designed to take up very little memory.

To use it, create a new text segmenter. By default, a model of word frequencies
from the web is loaded. Then call Segment() passing in some text. The return value
is the text split into strings containing individual words, unrecognized words, or
spaces and punctuation. You can get back the original input by concatenating the
results together.
*/
package chinese
