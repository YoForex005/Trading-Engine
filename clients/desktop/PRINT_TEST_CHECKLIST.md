# Print Functionality - QA Test Checklist

## Test Environment Setup

### Prerequisites
- [ ] Browser: Chrome/Edge (latest)
- [ ] Browser: Firefox (latest)
- [ ] Browser: Safari (latest)
- [ ] Printer: Physical printer connected OR PDF printer available
- [ ] Account: Test account with ID "1" or valid account
- [ ] Charts: Multiple charts with different symbols loaded

## Test Cases

### 1. Basic Print Functionality

#### TC-01: Quick Print via Keyboard
- [ ] Open application with chart loaded
- [ ] Press **Ctrl+P**
- [ ] Verify browser print dialog opens
- [ ] Verify chart preview shows in print dialog
- [ ] **Cancel** print
- [ ] Verify no errors in console (F12)

**Expected**: Print dialog opens with chart visible

#### TC-02: Print via File Menu
- [ ] Click **File** menu
- [ ] Click **Print**
- [ ] Verify browser print dialog opens
- [ ] Verify chart is captured correctly
- [ ] Cancel print
- [ ] Verify no errors

**Expected**: Same as TC-01

#### TC-03: Print with No Chart Loaded
- [ ] Close all charts
- [ ] Press **Ctrl+P**
- [ ] Verify behavior (error message or graceful handling)

**Expected**: Should show error or handle gracefully

---

### 2. Print Setup Dialog

#### TC-04: Open Print Setup
- [ ] Click **File** menu
- [ ] Click **Print Setup**
- [ ] Verify dialog opens
- [ ] Verify all controls are visible
- [ ] Verify default values loaded

**Expected**: Dialog opens with all options visible

#### TC-05: Change Orientation
- [ ] Open Print Setup
- [ ] Change to **Portrait**
- [ ] Change to **Landscape**
- [ ] Verify selection updates

**Expected**: Selection changes visible immediately

#### TC-06: Change Paper Size
- [ ] Open Print Setup
- [ ] Select **A4**
- [ ] Select **Letter**
- [ ] Select **Legal**
- [ ] Select **A3**
- [ ] Verify each selection

**Expected**: All paper sizes selectable

#### TC-07: Adjust Margins
- [ ] Open Print Setup
- [ ] Set Top margin to 10
- [ ] Set Right margin to 15
- [ ] Set Bottom margin to 10
- [ ] Set Left margin to 15
- [ ] Verify values accepted

**Expected**: All margin values between 0-50 accepted

#### TC-08: Margin Validation
- [ ] Try setting margin to -1 → Should prevent
- [ ] Try setting margin to 51 → Should prevent
- [ ] Try setting margin to 0 → Should allow
- [ ] Try setting margin to 50 → Should allow

**Expected**: Values outside 0-50 rejected

#### TC-09: Toggle Options
- [ ] Toggle **Include Header** ON/OFF
- [ ] Toggle **Include Footer** ON/OFF
- [ ] Toggle **Include Grid** ON/OFF
- [ ] Toggle **Include Indicators** ON/OFF
- [ ] Toggle **Include Drawings** ON/OFF
- [ ] Toggle **Scale to Fit** ON/OFF

**Expected**: All toggles work

#### TC-10: Custom Header Text
- [ ] Enable **Include Header**
- [ ] Enter custom text: "Test Header"
- [ ] Verify text accepted
- [ ] Clear text
- [ ] Verify default used

**Expected**: Custom text accepted and stored

#### TC-11: Custom Footer Text
- [ ] Enable **Include Footer**
- [ ] Enter custom text: "Test Footer"
- [ ] Verify text accepted
- [ ] Clear text
- [ ] Verify default used

**Expected**: Custom text accepted and stored

#### TC-12: Color Mode Selection
- [ ] Select **Color** mode
- [ ] Select **Black & White** mode
- [ ] Verify selection changes

**Expected**: Both modes selectable

#### TC-13: Save Preferences
- [ ] Configure custom settings
- [ ] Check **"Save as default"** checkbox
- [ ] Click **OK**
- [ ] Reload page
- [ ] Open Print Setup
- [ ] Verify settings persisted

**Expected**: Settings saved and loaded (requires backend API)

#### TC-14: Cancel Setup
- [ ] Open Print Setup
- [ ] Change some settings
- [ ] Click **Cancel**
- [ ] Open Print Setup again
- [ ] Verify changes not saved

**Expected**: Changes discarded on cancel

---

### 3. Print Preview Dialog

#### TC-15: Open Print Preview
- [ ] Click **File** → **Print Preview**
- [ ] Verify preview dialog opens
- [ ] Verify chart visible in preview
- [ ] Verify zoom controls visible

**Expected**: Preview opens with chart

#### TC-16: Zoom In
- [ ] Open Print Preview
- [ ] Click **Zoom In** button (+)
- [ ] Verify zoom increases (100% → 125% → 150%)
- [ ] Continue to maximum (200%)
- [ ] Verify button disabled at max

**Expected**: Zoom increases up to 200%, button disabled

#### TC-17: Zoom Out
- [ ] Open Print Preview
- [ ] Click **Zoom Out** button (-)
- [ ] Verify zoom decreases
- [ ] Continue to minimum (50%)
- [ ] Verify button disabled at min

**Expected**: Zoom decreases to 50%, button disabled

#### TC-18: Zoom Reset
- [ ] Zoom to 150%
- [ ] Click on percentage text (150%)
- [ ] Verify zoom resets to 100%

**Expected**: Zoom resets to 100%

#### TC-19: Print from Preview
- [ ] Open Print Preview
- [ ] Click **Print** button
- [ ] Verify browser print dialog opens
- [ ] Cancel print
- [ ] Verify preview closes

**Expected**: Print executes and preview closes

#### TC-20: Cancel Preview
- [ ] Open Print Preview
- [ ] Click **Cancel** or **X**
- [ ] Verify preview closes
- [ ] Verify no print executed

**Expected**: Preview closes without printing

#### TC-21: Preview with Custom Settings
- [ ] Open Print Setup
- [ ] Set Orientation to Portrait
- [ ] Set custom header text
- [ ] Click OK (opens preview)
- [ ] Verify preview shows portrait layout
- [ ] Verify header text visible

**Expected**: Preview reflects custom settings

---

### 4. Chart Capture

#### TC-22: Capture Simple Chart
- [ ] Load BTCUSD 1m chart
- [ ] Wait for chart to render
- [ ] Open Print Preview
- [ ] Verify chart canvas visible in preview

**Expected**: Chart image captured correctly

#### TC-23: Capture Chart with Indicators
- [ ] Add SMA(20) to chart
- [ ] Add RSI(14) to chart
- [ ] Open Print Setup
- [ ] Enable **Include Indicators**
- [ ] Open Preview
- [ ] Verify indicators listed

**Expected**: Indicators shown in legend

#### TC-24: Capture Chart with Drawings
- [ ] Draw trend line on chart
- [ ] Draw horizontal line on chart
- [ ] Open Print Setup
- [ ] Enable **Include Drawings**
- [ ] Open Preview
- [ ] Verify drawings listed

**Expected**: Drawings shown in legend

#### TC-25: Capture Different Chart Types
- [ ] Test with Candlestick chart
- [ ] Test with Line chart
- [ ] Test with Bar chart
- [ ] Verify all capture correctly

**Expected**: All chart types capture

#### TC-26: Capture Different Timeframes
- [ ] Test with 1m chart
- [ ] Test with 5m chart
- [ ] Test with 1h chart
- [ ] Test with 1d chart
- [ ] Verify all capture correctly

**Expected**: All timeframes capture

---

### 5. Print Output Verification

#### TC-27: Print to PDF - Landscape
- [ ] Open Print Setup
- [ ] Select Landscape orientation
- [ ] Select A4 paper
- [ ] Click OK
- [ ] Print to PDF
- [ ] Open PDF
- [ ] Verify chart is landscape orientation
- [ ] Verify chart fills page appropriately

**Expected**: Landscape PDF with chart

#### TC-28: Print to PDF - Portrait
- [ ] Open Print Setup
- [ ] Select Portrait orientation
- [ ] Print to PDF
- [ ] Verify chart is portrait orientation

**Expected**: Portrait PDF with chart

#### TC-29: Print with Header
- [ ] Enable header with custom text
- [ ] Print to PDF
- [ ] Open PDF
- [ ] Verify header at top
- [ ] Verify symbol, timeframe, date range shown
- [ ] Verify custom text shown

**Expected**: Header rendered correctly

#### TC-30: Print with Footer
- [ ] Enable footer with custom text
- [ ] Print to PDF
- [ ] Open PDF
- [ ] Verify footer at bottom
- [ ] Verify custom text shown
- [ ] Verify page number shown

**Expected**: Footer rendered correctly

#### TC-31: Print in Color
- [ ] Set Color mode
- [ ] Print to PDF
- [ ] Open PDF
- [ ] Verify chart is in color

**Expected**: Full color output

#### TC-32: Print in Grayscale
- [ ] Set Black & White mode
- [ ] Print to PDF
- [ ] Open PDF
- [ ] Verify chart is grayscale

**Expected**: Grayscale output

#### TC-33: Print with Legends
- [ ] Enable all legends
- [ ] Add indicators and drawings
- [ ] Print to PDF
- [ ] Verify legends appear in PDF
- [ ] Verify legends are readable

**Expected**: Legends visible and readable

#### TC-34: Print with Large Margins
- [ ] Set all margins to 40mm
- [ ] Print to PDF
- [ ] Verify large margins
- [ ] Verify chart still visible

**Expected**: Large margins applied

#### TC-35: Print with Small Margins
- [ ] Set all margins to 5mm
- [ ] Print to PDF
- [ ] Verify small margins
- [ ] Verify chart fills more of page

**Expected**: Small margins applied

---

### 6. Edge Cases

#### TC-36: Print Very Long Chart
- [ ] Load chart with 1000+ candles
- [ ] Print to PDF
- [ ] Verify chart doesn't overflow
- [ ] Verify scale-to-fit works

**Expected**: Chart fits on page

#### TC-37: Print Chart with Many Indicators
- [ ] Add 5+ indicators
- [ ] Enable indicator legend
- [ ] Print to PDF
- [ ] Verify all indicators listed

**Expected**: All indicators shown (may truncate if too many)

#### TC-38: Print Empty Chart
- [ ] Load symbol with no data
- [ ] Try to print
- [ ] Verify behavior

**Expected**: Handles gracefully (empty chart or error)

#### TC-39: Print During Chart Update
- [ ] Start print preview
- [ ] While preview open, new tick arrives
- [ ] Verify preview not affected

**Expected**: Preview shows snapshot, not live data

#### TC-40: Rapid Print Requests
- [ ] Press Ctrl+P
- [ ] Immediately press Ctrl+P again
- [ ] Verify no errors or conflicts

**Expected**: Handles gracefully (queue or ignore)

---

### 7. Browser Compatibility

#### TC-41: Chrome/Edge Print
- [ ] Test all print functions in Chrome
- [ ] Verify print dialog appearance
- [ ] Verify chart capture quality
- [ ] Print to PDF and verify output

**Expected**: Full functionality in Chrome/Edge

#### TC-42: Firefox Print
- [ ] Test all print functions in Firefox
- [ ] Verify print dialog appearance
- [ ] Verify chart capture quality
- [ ] Print to PDF and verify output

**Expected**: Full functionality in Firefox

#### TC-43: Safari Print
- [ ] Test all print functions in Safari (Mac)
- [ ] Verify print dialog appearance
- [ ] Verify chart capture quality
- [ ] Print to PDF and verify output

**Expected**: Full functionality in Safari

---

### 8. Error Handling

#### TC-44: Print with No Printer
- [ ] Disable/disconnect all printers
- [ ] Try to print
- [ ] Verify error handling

**Expected**: Browser handles (not app error)

#### TC-45: Cancel Print Dialog
- [ ] Start print
- [ ] Cancel browser print dialog
- [ ] Verify app continues normally

**Expected**: Cancellation handled gracefully

#### TC-46: Invalid Preferences
- [ ] Manually corrupt preferences in localStorage
- [ ] Try to load Print Setup
- [ ] Verify falls back to defaults

**Expected**: Uses defaults if preferences invalid

---

### 9. Performance

#### TC-47: Print Response Time
- [ ] Press Ctrl+P
- [ ] Measure time to print dialog
- [ ] Should be < 2 seconds

**Expected**: Fast response (< 2s)

#### TC-48: Preview Load Time
- [ ] Open Print Preview
- [ ] Measure time to see preview
- [ ] Should be < 3 seconds

**Expected**: Preview loads quickly (< 3s)

#### TC-49: Memory Leak Check
- [ ] Open/close Print Preview 20 times
- [ ] Check browser memory usage
- [ ] Verify no significant increase

**Expected**: No memory leaks

---

### 10. Integration

#### TC-50: Print from Chart Context Menu
- [ ] Right-click on chart
- [ ] Select Print (if available)
- [ ] Verify print executes

**Expected**: Context menu print works (if implemented)

#### TC-51: Print After Chart Switch
- [ ] Open multiple chart tabs
- [ ] Switch between tabs
- [ ] Print from each tab
- [ ] Verify correct chart prints

**Expected**: Active chart prints

#### TC-52: Print After Timeframe Change
- [ ] Change timeframe
- [ ] Immediately print
- [ ] Verify new timeframe captured

**Expected**: Current timeframe prints

---

## Defect Severity Guidelines

### Critical (P0)
- Print crashes application
- Data loss or corruption
- Security vulnerability

### High (P1)
- Print produces completely blank page
- Chart not captured at all
- Keyboard shortcut doesn't work

### Medium (P2)
- Minor visual issues in output
- Legends not rendering correctly
- Preferences not saving

### Low (P3)
- Cosmetic issues
- Minor text alignment problems
- Nice-to-have features missing

---

## Test Results Summary

### Passed: _____ / 52
### Failed: _____ / 52
### Blocked: _____ / 52
### Not Tested: _____ / 52

---

## Sign-off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| QA Lead | | | |
| Developer | | | |
| Product Owner | | | |

---

**Notes**:
- Some tests require backend API implementation (preference persistence)
- PDF printing is browser-dependent (use Print to PDF option)
- Physical printer tests optional but recommended
- Test in private/incognito mode to verify clean state

**Version**: 1.0
**Last Updated**: 2026-02-02
