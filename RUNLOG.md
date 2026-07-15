| Profile | delay_ms | Miss % | Overhead | Changes & Rationale |
| :--- | :--- | :--- | :--- | :--- |
| A | 40ms | 100.0% | 1.00x | Baseline unmodified sender/receiver. Fails completely because normal network delay exceeds the 40ms deadline, causing almost every packet to miss. |
| B | 40ms | 100.0% | 1.00x | Baseline unmodified. Fails for the same reason. |
| A | 100ms | 0.40% | 1.90x | Introduced k=3 FEC with an artificial 1.90x budget limit to save headroom for ARQ. Miss rate passed perfectly, but the artificial budget limit caused an unnecessary 5% hole in FEC protection. |
| B | 140ms | 1.47% | 1.90x | Same artificial 1.90x FEC limit. Miss rate failed (>1%) because the 5% hole in FEC protection multiplied with Profile B's 5% independent loss rate caused too many permanent losses. |
| A | 100ms | 0.27% | 2.00x | Removed artificial 1.90x budget. Target full 2.00x dynamically, only dropping FEC when ARQ actively steals the bandwidth. Passes flawlessly. |
| B | 140ms | 1.00% | 2.00x | Full 2.00x dynamic budget. FEC provides complete coverage. Passes flawlessly. |
| D_burst3 | 300ms | 0.73% | 1.90x | Tested on a custom extreme burst-loss profile (`D_burst3.json`) to force ARQ fallback. `delay_ms` boosted to 300 to allow NACK round trips. Passes flawlessly, proving the hybrid ARQ protocol rescues massive bursts. |
| A | 140ms | 0.20% | 2.00x | Final validation using the chosen 140ms delay. Memory optimizations applied (zero-allocation arrays replacing maps and channels). Passes perfectly. |
| B | 140ms | 0.53% | 2.00x | Final validation using the chosen 140ms delay. Passes perfectly. |
