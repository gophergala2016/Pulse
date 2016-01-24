/*Package pulse uses a pattern identification algorithm created by Michael Dropps.
The custom maching-learning algorithm identifies patterns that it finds in strings.
This is better to use than Levenshtein distance because the unique features of input strings are
stored and available for future lookups.  It is a more advanced approach than a simple distance comparison.
The Levenshtein distance is used on first-pass comparisons to help the algorithm create initial patterns
using the custom matrix-based algorithm.  The inputs with the most similarities are compared first until
a few patterns are in memory.  A custom hashing approach is used to hash the unique aspects of each pattern
into a constrained array of maps.  By using the this map as a lookup table, we are easily able to detect
any existing pattern that is likely to match the input.  If this lookup does not present a pattern, we again
fall back to the Levenstein distance to compare the input against all unmatched inputs, just as in the beginning.
By following this order, new patterns are created only when they are not good matches to existing patterns.
It will create patterns it finds and compare incoming strings to them.
If it is a close match, the current pattern will likely be altered to account for the new information.
This algorithm is always learning new patterns and revising existing patterns, according to the input.
If a string doesn't match any pattern, it is put into an unmatched state.
Unmatched strings that remain unmatched after a certain period of time are reported
as an anomaly and are sent to the user using the function supplied by the consuming routine at startup.
Using this method, the user can do anything they want with the anomalies found.
*/
package pulse
