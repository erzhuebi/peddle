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

peddle_net_cmd_at:
    .byte 65, 84, 13

; ATDT
peddle_net_cmd_atdt:
    .byte 65, 84, 68, 84

peddle_net_resp_ok:
    .byte 79, 75

peddle_net_resp_connect:
    .byte 67, 79, 78, 78, 69, 67, 84

peddle_acia_init:
    lda #$1f
    sta ACIA_CONTROL
    lda #$0b
    sta ACIA_COMMAND
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

peddle_netconnect:
    jsr peddle_acia_init
    jsr peddle_acia_flush

    lda #0
    sta peddle_net_connected

    lda #<peddle_net_cmd_at
    sta ZP_PTR1_LO
    lda #>peddle_net_cmd_at
    sta ZP_PTR1_HI
    lda #3
    sta peddle_net_limit_lo
    lda #0
    sta peddle_net_limit_hi
    jsr peddle_net_send_raw

    lda #<peddle_net_resp_ok
    sta peddle_net_pattern_lo
    lda #>peddle_net_resp_ok
    sta peddle_net_pattern_hi
    lda #2
    sta peddle_net_pattern_len
    lda #100
    sta peddle_net_timeout_lo
    lda #0
    sta peddle_net_timeout_hi
    jsr peddle_net_expect
    cmp #0
    beq peddle_netconnect_fail

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

    lda #<peddle_net_resp_connect
    sta peddle_net_pattern_lo
    lda #>peddle_net_resp_connect
    sta peddle_net_pattern_hi
    lda #7
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
    lda peddle_net_count_lo
    ora peddle_net_count_hi
    bne peddle_netread_done

    lda peddle_net_timeout_lo
    ora peddle_net_timeout_hi
    beq peddle_netread_done

    jsr peddle_net_timeout_due
    cmp #0
    beq peddle_netread_loop

peddle_netread_done:
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

peddle_netclose:
    jsr peddle_acia_drop_dtr
    lda #0
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

    ldy #0
    lda (ZP_PTR0_LO), y
    sta peddle_net_limit_lo
    iny
    lda (ZP_PTR0_LO), y
    sta peddle_net_limit_hi

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
