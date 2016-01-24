/*Package pulse uses an algorithm creted by Michael Dropps.
The algorithm learns from patterns that it finds in strings.
This is better to use than Levenshtein distance because string length is not affected as much when comparing strings.
The Levenshtein distance is used on first compare to help this algorithm out put it goes into more details.
It will create patterns it finds and compare incoming strings to them.
If it is a close match the current pattern may be altered to compensate for a match.
This algorithm is always learning new patterns.
If a string doesn't match any pattern it is put into an unmatched state.
After a while it says it is an anomaly and sends it back to the user on the function specified.
That way the user can do anything they want with anomalies found.
*/
package pulse
