package codegen

func (g *Generator) emitNetRuntime() {
	if !g.usedNetRuntime {
		return
	}

	g.emit(`
; network runtime
;
; C64 Ultimate modem simulator / SwiftLink-style 6551 ACIA.
; Default base address:
;   $DE00 data
;   $DE01 status
;   $DE02 command
;   $DE03 control
;
; Status bits used:
;   bit 3 = receiver data register full
;   bit 4 = transmitter data register empty
;
; Public helpers emitted by codegen:
;   peddle_netconnect
;   peddle_netread
;   peddle_netreadlf
;   peddle_netwrite
;   peddle_netclose
;
; Peddle API semantics:
;   netconnect(addr char[], port int) bool
;   netread(buffer byte[]|char[], max int, timeoutTicks int) int
;   netreadlf(buffer byte[]|char[], max int, timeoutTicks int) bool
;   netwrite(buffer byte[]|char[], len int) int
;   netclose()
;   netconnected() bool
;
; Notes:
;   - Single global connection only.
;   - Timeout is measured in C64 ticks/jiffies.
;   - Uses ATDT because that matches the known-good C64U test program.
;   - Mirrors the working C test:
;       init ACIA
;       send +++
;       send ATH<CR>
;       send ATDT<addr>:<port><CR>
;       wait for CONNECT
;       enter data mode without flushing TCP payload

ACIA_DATA    = $de00
ACIA_STATUS  = $de01
ACIA_COMMAND = $de02
ACIA_CONTROL = $de03

peddle_net_connected:
    .byte 0

peddle_net_addr_lo:
    .byte 0
peddle_net_addr_hi:
    .byte 0

peddle_net_buf_lo:
    .byte 0
peddle_net_buf_hi:
    .byte 0

peddle_net_max_lo:
    .byte 0
peddle_net_max_hi:
    .byte 0

peddle_net_timeout_lo:
    .byte 0
peddle_net_timeout_hi:
    .byte 0

peddle_net_limit_lo:
    .byte 0
peddle_net_limit_hi:
    .byte 0

peddle_net_count_lo:
    .byte 0
peddle_net_count_hi:
    .byte 0

peddle_net_start_lo:
    .byte 0
peddle_net_start_hi:
    .byte 0

peddle_net_pattern_lo:
    .byte 0
peddle_net_pattern_hi:
    .byte 0

peddle_net_pattern_len:
    .byte 0

peddle_net_match_index:
    .byte 0

peddle_net_last_char:
    .byte 0

peddle_net_line_found:
    .byte 0

peddle_net_had_byte:
    .byte 0

peddle_net_skip_lf:
    .byte 0

peddle_net_port_lo:
    .byte 0
peddle_net_port_hi:
    .byte 0

peddle_net_div_lo:
    .byte 0
peddle_net_div_hi:
    .byte 0

peddle_net_digit:
    .byte 0

peddle_net_started:
    .byte 0

peddle_net_force:
    .byte 0

; "+++"
peddle_net_cmd_escape:
    .byte 43, 43, 43

; "ATH\r"
peddle_net_cmd_ath:
    .byte 65, 84, 72, 13

; "ATDT"
peddle_net_cmd_atdt:
    .byte 65, 84, 68, 84

; "CONNECT\r"
peddle_net_resp_connect:
    .byte 67, 79, 78, 78, 69, 67, 84, 13

peddle_acia_init:
    lda #$0b
    sta ACIA_COMMAND
    lda #$1f
    sta ACIA_CONTROL
    rts

peddle_acia_drop_dtr:
    lda #$00
    sta ACIA_COMMAND
    rts

peddle_acia_can_read:
    lda ACIA_STATUS
    and #$08
    rts

peddle_acia_can_write:
    lda ACIA_STATUS
    and #$10
    rts

peddle_acia_read:
    lda ACIA_DATA
    rts

peddle_acia_write:
    pha

peddle_acia_write_wait:
    lda ACIA_STATUS
    and #$10
    beq peddle_acia_write_wait
    pla
    sta ACIA_DATA
    rts

peddle_acia_flush:
    jsr peddle_acia_can_read
    beq peddle_acia_flush_done
    lda ACIA_DATA
    jmp peddle_acia_flush

peddle_acia_flush_done:
    rts

peddle_net_guard_delay:
    lda #0
    sta peddle_net_count_lo
    sta peddle_net_count_hi

peddle_net_guard_delay_loop:
    inc peddle_net_count_lo
    bne peddle_net_guard_delay_check
    inc peddle_net_count_hi

peddle_net_guard_delay_check:
    lda peddle_net_count_hi
    cmp #4
    bcc peddle_net_guard_delay_loop
    rts

peddle_netconnect:
    jsr peddle_acia_init
    jsr peddle_acia_flush

    lda #0
    sta peddle_net_connected

    ; Match the known-good C test:
    ; guard delay, "+++", guard delay, "ATH\r", guard delay.
    jsr peddle_net_guard_delay

    lda #<peddle_net_cmd_escape
    sta ZP_PTR1_LO
    lda #>peddle_net_cmd_escape
    sta ZP_PTR1_HI
    lda #3
    sta peddle_net_limit_lo
    lda #0
    sta peddle_net_limit_hi
    jsr peddle_net_send_raw

    jsr peddle_net_guard_delay

    lda #<peddle_net_cmd_ath
    sta ZP_PTR1_LO
    lda #>peddle_net_cmd_ath
    sta ZP_PTR1_HI
    lda #4
    sta peddle_net_limit_lo
    lda #0
    sta peddle_net_limit_hi
    jsr peddle_net_send_raw

    jsr peddle_net_guard_delay
    jsr peddle_acia_flush

    ; Send "ATDT" + addr + ":" + port + "\r".
    lda #<peddle_net_cmd_atdt
    sta ZP_PTR1_LO
    lda #>peddle_net_cmd_atdt
    sta ZP_PTR1_HI
    lda #4
    sta peddle_net_limit_lo
    lda #0
    sta peddle_net_limit_hi
    jsr peddle_net_send_raw

    jsr peddle_net_send_addr

    lda #58
    jsr peddle_acia_write

    jsr peddle_net_send_port_decimal

    lda #13
    jsr peddle_acia_write

    ; Wait for the complete modem result line terminator. The first byte
    ; after this match belongs to TCP payload and must not be flushed.
    lda #<peddle_net_resp_connect
    sta peddle_net_pattern_lo
    lda #>peddle_net_resp_connect
    sta peddle_net_pattern_hi
    lda #8
    sta peddle_net_pattern_len
    lda #<600
    sta peddle_net_timeout_lo
    lda #>600
    sta peddle_net_timeout_hi
    jsr peddle_net_expect
    cmp #0
    beq peddle_netconnect_fail

    lda #1
    sta peddle_net_connected
    lda #0
    sta peddle_net_skip_lf
    lda #1
    rts

peddle_netconnect_fail:
    lda #0
    sta peddle_net_connected
    rts

peddle_net_send_addr:
    lda peddle_net_addr_lo
    sta ZP_PTR0_LO
    lda peddle_net_addr_hi
    sta ZP_PTR0_HI

    ; char[] layout:
    ;   +0/+1 capacity
    ;   +2/+3 length
    ;   +4    data
    ldy #2
    lda (ZP_PTR0_LO), y
    sta peddle_net_limit_lo
    iny
    lda (ZP_PTR0_LO), y
    sta peddle_net_limit_hi

    lda peddle_net_addr_lo
    clc
    adc #4
    sta ZP_PTR1_LO
    lda peddle_net_addr_hi
    adc #0
    sta ZP_PTR1_HI

    jmp peddle_net_send_raw

peddle_net_send_raw:
    lda #0
    sta peddle_net_count_lo
    sta peddle_net_count_hi

peddle_net_send_raw_loop:
    jsr peddle_net_count_reached_limit
    bne peddle_net_send_raw_done

    ldy #0
    lda (ZP_PTR1_LO), y
    jsr peddle_acia_write

    jsr peddle_net_inc_data_ptr
    jsr peddle_net_inc_count
    jmp peddle_net_send_raw_loop

peddle_net_send_raw_done:
    rts

peddle_net_send_port_decimal:
    lda #0
    sta peddle_net_started

    lda #<10000
    sta peddle_net_div_lo
    lda #>10000
    sta peddle_net_div_hi
    lda #0
    sta peddle_net_force
    jsr peddle_net_send_port_digit

    lda #<1000
    sta peddle_net_div_lo
    lda #>1000
    sta peddle_net_div_hi
    lda #0
    sta peddle_net_force
    jsr peddle_net_send_port_digit

    lda #<100
    sta peddle_net_div_lo
    lda #>100
    sta peddle_net_div_hi
    lda #0
    sta peddle_net_force
    jsr peddle_net_send_port_digit

    lda #<10
    sta peddle_net_div_lo
    lda #>10
    sta peddle_net_div_hi
    lda #0
    sta peddle_net_force
    jsr peddle_net_send_port_digit

    lda #<1
    sta peddle_net_div_lo
    lda #>1
    sta peddle_net_div_hi
    lda #1
    sta peddle_net_force
    jsr peddle_net_send_port_digit

    rts

peddle_net_send_port_digit:
    lda #0
    sta peddle_net_digit

peddle_net_port_digit_loop:
    lda peddle_net_port_hi
    cmp peddle_net_div_hi
    bcc peddle_net_port_digit_done
    bne peddle_net_port_digit_subtract

    lda peddle_net_port_lo
    cmp peddle_net_div_lo
    bcc peddle_net_port_digit_done

peddle_net_port_digit_subtract:
    sec
    lda peddle_net_port_lo
    sbc peddle_net_div_lo
    sta peddle_net_port_lo
    lda peddle_net_port_hi
    sbc peddle_net_div_hi
    sta peddle_net_port_hi

    inc peddle_net_digit
    jmp peddle_net_port_digit_loop

peddle_net_port_digit_done:
    lda peddle_net_force
    bne peddle_net_port_digit_emit

    lda peddle_net_started
    bne peddle_net_port_digit_emit

    lda peddle_net_digit
    bne peddle_net_port_digit_start

    rts

peddle_net_port_digit_start:
    lda #1
    sta peddle_net_started

peddle_net_port_digit_emit:
    lda peddle_net_digit
    clc
    adc #48
    jsr peddle_acia_write
    rts

peddle_net_expect:
    ; 6502 indexed indirect load requires a zero-page pointer.
    ; peddle_net_pattern_lo/hi are normal labels, so copy them to ZP_PTR0 first.
    lda peddle_net_pattern_lo
    sta ZP_PTR0_LO
    lda peddle_net_pattern_hi
    sta ZP_PTR0_HI

    lda #0
    sta peddle_net_match_index

    lda $00a2
    sta peddle_net_start_lo
    lda $00a1
    sta peddle_net_start_hi

peddle_net_expect_loop:
    jsr peddle_acia_can_read
    beq peddle_net_expect_no_byte

    jsr peddle_acia_read
    sta peddle_net_last_char

    ldy peddle_net_match_index
    lda (ZP_PTR0_LO), y
    cmp peddle_net_last_char
    beq peddle_net_expect_match

    lda #0
    sta peddle_net_match_index

    ldy #0
    lda (ZP_PTR0_LO), y
    cmp peddle_net_last_char
    bne peddle_net_expect_no_byte

    lda #1
    sta peddle_net_match_index
    lda peddle_net_pattern_len
    cmp #1
    beq peddle_net_expect_success
    jmp peddle_net_expect_no_byte

peddle_net_expect_match:
    inc peddle_net_match_index
    lda peddle_net_match_index
    cmp peddle_net_pattern_len
    beq peddle_net_expect_success

peddle_net_expect_no_byte:
    jsr peddle_net_timeout_due
    cmp #0
    beq peddle_net_expect_loop

    lda #0
    rts

peddle_net_expect_success:
    lda #1
    rts

peddle_netread:
    jsr peddle_net_prepare_buffer
    jsr peddle_net_limit_capacity_max

    lda #0
    sta peddle_net_count_lo
    sta peddle_net_count_hi

    ; Clear destination array length.
    lda peddle_net_buf_lo
    sta ZP_PTR0_LO
    lda peddle_net_buf_hi
    sta ZP_PTR0_HI

    ldy #2
    lda #0
    sta (ZP_PTR0_LO), y
    iny
    sta (ZP_PTR0_LO), y

    lda $00a2
    sta peddle_net_start_lo
    lda $00a1
    sta peddle_net_start_hi

peddle_netread_loop:
    jsr peddle_net_count_reached_limit
    bne peddle_netread_done

    jsr peddle_acia_can_read
    beq peddle_netread_no_byte

    jsr peddle_acia_read

    ldy #0
    sta (ZP_PTR1_LO), y

    jsr peddle_net_inc_data_ptr
    jsr peddle_net_inc_count

    jmp peddle_netread_loop

peddle_netread_no_byte:
    ; If at least one byte was read, return immediately with the available
    ; chunk. Timeout only waits for the first byte.
    lda peddle_net_count_lo
    ora peddle_net_count_hi
    bne peddle_netread_done

    ; timeout 0 means non-blocking.
    lda peddle_net_timeout_lo
    ora peddle_net_timeout_hi
    beq peddle_netread_done

    jsr peddle_net_timeout_due
    cmp #0
    beq peddle_netread_loop

peddle_netread_done:
    ; Store resulting length into destination array.
    lda peddle_net_buf_lo
    sta ZP_PTR0_LO
    lda peddle_net_buf_hi
    sta ZP_PTR0_HI

    ldy #2
    lda peddle_net_count_lo
    sta (ZP_PTR0_LO), y
    iny
    lda peddle_net_count_hi
    sta (ZP_PTR0_LO), y

    lda peddle_net_count_lo
    sta ZP_TMP0
    lda peddle_net_count_hi
    sta ZP_TMP1
    lda ZP_TMP0
    rts

peddle_netwrite:
    jsr peddle_net_prepare_buffer
    jsr peddle_net_limit_capacity_max

    lda #0
    sta peddle_net_count_lo
    sta peddle_net_count_hi

peddle_netwrite_loop:
    jsr peddle_net_count_reached_limit
    bne peddle_netwrite_done

    ldy #0
    lda (ZP_PTR1_LO), y
    jsr peddle_acia_write

    jsr peddle_net_inc_data_ptr
    jsr peddle_net_inc_count
    jmp peddle_netwrite_loop

peddle_netwrite_done:
    lda peddle_net_count_lo
    sta ZP_TMP0
    lda peddle_net_count_hi
    sta ZP_TMP1
    lda ZP_TMP0
    rts

peddle_netreadlf:
    jsr peddle_net_prepare_buffer
    jsr peddle_net_limit_capacity_max

    lda #0
    sta peddle_net_line_found
    sta peddle_net_had_byte

    ; Start with the existing destination length. netreadlf() appends into
    ; the caller-owned line buffer and excludes CR/LF terminators.
    lda peddle_net_buf_lo
    sta ZP_PTR0_LO
    lda peddle_net_buf_hi
    sta ZP_PTR0_HI

    ldy #2
    lda (ZP_PTR0_LO), y
    sta peddle_net_count_lo
    iny
    lda (ZP_PTR0_LO), y
    sta peddle_net_count_hi

    lda ZP_PTR1_LO
    clc
    adc peddle_net_count_lo
    sta ZP_PTR1_LO
    lda ZP_PTR1_HI
    adc peddle_net_count_hi
    sta ZP_PTR1_HI

    lda $00a2
    sta peddle_net_start_lo
    lda $00a1
    sta peddle_net_start_hi

peddle_netreadlf_loop:
    jsr peddle_net_count_reached_limit
    bne peddle_netreadlf_done

    jsr peddle_acia_can_read
    beq peddle_netreadlf_no_byte

    lda #1
    sta peddle_net_had_byte

    jsr peddle_acia_read
    sta peddle_net_last_char

    ; If the previous line ended with CR, swallow one following LF so CRLF
    ; does not become an empty next line.
    lda peddle_net_skip_lf
    beq peddle_netreadlf_check_terminator

    lda #0
    sta peddle_net_skip_lf

    lda peddle_net_last_char
    cmp #10
    beq peddle_netreadlf_loop

peddle_netreadlf_check_terminator:
    lda peddle_net_last_char
    cmp #13
    beq peddle_netreadlf_found_cr

    cmp #10
    beq peddle_netreadlf_found_lf

    ldy #0
    sta (ZP_PTR1_LO), y

    jsr peddle_net_inc_data_ptr
    jsr peddle_net_inc_count

    jmp peddle_netreadlf_loop

peddle_netreadlf_found_cr:
    lda #1
    sta peddle_net_skip_lf
    sta peddle_net_line_found
    jmp peddle_netreadlf_done

peddle_netreadlf_found_lf:
    lda #1
    sta peddle_net_line_found
    jmp peddle_netreadlf_done

peddle_netreadlf_no_byte:
    ; If any byte was read or consumed, return immediately. Timeout only
    ; waits for the first byte, just like netread().
    lda peddle_net_had_byte
    bne peddle_netreadlf_done

    ; timeout 0 means non-blocking.
    lda peddle_net_timeout_lo
    ora peddle_net_timeout_hi
    beq peddle_netreadlf_done

    jsr peddle_net_timeout_due
    cmp #0
    beq peddle_netreadlf_loop

peddle_netreadlf_done:
    ; Store resulting length into destination array.
    lda peddle_net_buf_lo
    sta ZP_PTR0_LO
    lda peddle_net_buf_hi
    sta ZP_PTR0_HI

    ldy #2
    lda peddle_net_count_lo
    sta (ZP_PTR0_LO), y
    iny
    lda peddle_net_count_hi
    sta (ZP_PTR0_LO), y

    lda peddle_net_line_found
    rts

peddle_netclose:
    jsr peddle_net_guard_delay

    lda #<peddle_net_cmd_escape
    sta ZP_PTR1_LO
    lda #>peddle_net_cmd_escape
    sta ZP_PTR1_HI
    lda #3
    sta peddle_net_limit_lo
    lda #0
    sta peddle_net_limit_hi
    jsr peddle_net_send_raw

    jsr peddle_net_guard_delay

    lda #<peddle_net_cmd_ath
    sta ZP_PTR1_LO
    lda #>peddle_net_cmd_ath
    sta ZP_PTR1_HI
    lda #4
    sta peddle_net_limit_lo
    lda #0
    sta peddle_net_limit_hi
    jsr peddle_net_send_raw

    jsr peddle_net_guard_delay
    jsr peddle_acia_drop_dtr

    lda #0
    sta peddle_net_skip_lf
    sta peddle_net_connected
    rts

peddle_net_prepare_buffer:
    lda peddle_net_buf_lo
    clc
    adc #4
    sta ZP_PTR1_LO
    lda peddle_net_buf_hi
    adc #0
    sta ZP_PTR1_HI
    rts

peddle_net_limit_capacity_max:
    lda peddle_net_buf_lo
    sta ZP_PTR0_LO
    lda peddle_net_buf_hi
    sta ZP_PTR0_HI

    ; limit = array capacity
    ldy #0
    lda (ZP_PTR0_LO), y
    sta peddle_net_limit_lo
    iny
    lda (ZP_PTR0_LO), y
    sta peddle_net_limit_hi

    ; if capacity <= max, keep capacity
    lda peddle_net_limit_hi
    cmp peddle_net_max_hi
    bcc peddle_net_limit_done
    bne peddle_net_use_max

    lda peddle_net_limit_lo
    cmp peddle_net_max_lo
    bcc peddle_net_limit_done
    beq peddle_net_limit_done

peddle_net_use_max:
    lda peddle_net_max_lo
    sta peddle_net_limit_lo
    lda peddle_net_max_hi
    sta peddle_net_limit_hi

peddle_net_limit_done:
    rts

peddle_net_count_reached_limit:
    lda peddle_net_count_hi
    cmp peddle_net_limit_hi
    bcc peddle_net_count_not_reached
    bne peddle_net_count_reached

    lda peddle_net_count_lo
    cmp peddle_net_limit_lo
    bcc peddle_net_count_not_reached

peddle_net_count_reached:
    lda #1
    rts

peddle_net_count_not_reached:
    lda #0
    rts

peddle_net_inc_data_ptr:
    inc ZP_PTR1_LO
    bne peddle_net_inc_data_ptr_done
    inc ZP_PTR1_HI

peddle_net_inc_data_ptr_done:
    rts

peddle_net_inc_count:
    inc peddle_net_count_lo
    bne peddle_net_inc_count_done
    inc peddle_net_count_hi

peddle_net_inc_count_done:
    rts

peddle_net_timeout_due:
    ; timeout 0 means no wait / immediately due.
    lda peddle_net_timeout_lo
    ora peddle_net_timeout_hi
    bne peddle_net_timeout_nonzero

    lda #1
    rts

peddle_net_timeout_nonzero:
    ; elapsed = current ticks - start ticks, wrap-safe for normal intervals.
    ; KERNAL jiffy clock low/high:
    ;   low  = $A2
    ;   high = $A1
    lda $00a2
    sec
    sbc peddle_net_start_lo
    sta ZP_TMP0
    lda $00a1
    sbc peddle_net_start_hi
    sta ZP_TMP1

    lda ZP_TMP1
    cmp peddle_net_timeout_hi
    bcc peddle_net_timeout_not_due
    bne peddle_net_timeout_is_due

    lda ZP_TMP0
    cmp peddle_net_timeout_lo
    bcc peddle_net_timeout_not_due

peddle_net_timeout_is_due:
    lda #1
    rts

peddle_net_timeout_not_due:
    lda #0
    rts
`)
}
