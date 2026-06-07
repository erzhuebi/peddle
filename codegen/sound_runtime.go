package codegen

func (g *Generator) emitSoundRuntime() {
	if !g.usedSoundRuntime {
		return
	}

	g.emit(`
; sound runtime
;
; API:
;   sound_init(pool byte[])
;   sound_reset()
;   sound_load(data byte[], kind int) (uint, int)
;   sound_play(id)
;   sound_stop(id)
;   sound_num() int
;   sound_memfree() int
;
; V1 sound kind:
;   1 = SID register stream
;
; Register stream commands:
;   0               end
;   1, frames       wait frames IRQ ticks
;   2, reg, value   write value to $D400 + reg

PEDDLE_SOUND_REGSTREAM = 1
PEDDLE_SOUND_MAX       = 16

peddle_sound_load_return_id:
    .fill 2, 0
peddle_sound_load_return_err:
    .fill 2, 0

peddle_sound_initialized:
    .byte 0
peddle_sound_irq_installed:
    .byte 0

; Avoid the NMOS 6502 JMP ($xxff) indirect-vector bug when chaining IRQs.
.if (* & $00ff) == $00ff
    .byte 0
.endif
peddle_sound_old_irq_lo:
    .byte 0
peddle_sound_old_irq_hi:
    .byte 0

peddle_sound_pool_header_lo:
    .byte 0
peddle_sound_pool_header_hi:
    .byte 0
peddle_sound_pool_data_lo:
    .byte 0
peddle_sound_pool_data_hi:
    .byte 0
peddle_sound_pool_cap_lo:
    .byte 0
peddle_sound_pool_cap_hi:
    .byte 0
peddle_sound_pool_used_lo:
    .byte 0
peddle_sound_pool_used_hi:
    .byte 0
peddle_sound_new_used_lo:
    .byte 0
peddle_sound_new_used_hi:
    .byte 0

peddle_sound_data_header_lo:
    .byte 0
peddle_sound_data_header_hi:
    .byte 0
peddle_sound_data_lo:
    .byte 0
peddle_sound_data_hi:
    .byte 0
peddle_sound_data_len_lo:
    .byte 0
peddle_sound_data_len_hi:
    .byte 0

peddle_sound_kind_lo:
    .byte 0
peddle_sound_kind_hi:
    .byte 0
peddle_sound_handle_lo:
    .byte 0
peddle_sound_handle_hi:
    .byte 0

peddle_sound_slot_index:
    .byte 0
peddle_sound_result_id:
    .byte 0
peddle_sound_count:
    .byte 0

peddle_sound_copy_remaining_lo:
    .byte 0
peddle_sound_copy_remaining_hi:
    .byte 0
peddle_sound_validate_remaining_lo:
    .byte 0
peddle_sound_validate_remaining_hi:
    .byte 0

peddle_sound_playing:
    .byte 0
peddle_sound_active_id:
    .byte 0
peddle_sound_wait:
    .byte 0
peddle_sound_pc_lo:
    .byte 0
peddle_sound_pc_hi:
    .byte 0

peddle_sound_irq_save_ptr0_lo:
    .byte 0
peddle_sound_irq_save_ptr0_hi:
    .byte 0

peddle_sound_slot_inuse:
    .fill 16, 0
peddle_sound_slot_kind:
    .fill 16, 0
peddle_sound_slot_offset_lo:
    .fill 16, 0
peddle_sound_slot_offset_hi:
    .fill 16, 0
peddle_sound_slot_len_lo:
    .fill 16, 0
peddle_sound_slot_len_hi:
    .fill 16, 0

peddle_sound_init:
    lda #1
    sta peddle_sound_initialized

    jsr peddle_sound_install_irq
    jsr peddle_sound_setup_pool
    jsr peddle_sound_reset
    rts

peddle_sound_setup_pool:
    lda peddle_sound_pool_header_lo
    sta ZP_PTR0_LO
    lda peddle_sound_pool_header_hi
    sta ZP_PTR0_HI

    ldy #0
    lda (ZP_PTR0_LO), y
    sta peddle_sound_pool_cap_lo
    iny
    lda (ZP_PTR0_LO), y
    sta peddle_sound_pool_cap_hi

    lda peddle_sound_pool_header_lo
    clc
    adc #4
    sta peddle_sound_pool_data_lo
    lda peddle_sound_pool_header_hi
    adc #0
    sta peddle_sound_pool_data_hi
    rts

peddle_sound_install_irq:
    lda peddle_sound_irq_installed
    beq peddle_sound_install_irq_now
    rts

peddle_sound_install_irq_now:
    sei
    lda $0314
    sta peddle_sound_old_irq_lo
    lda $0315
    sta peddle_sound_old_irq_hi
    lda #<peddle_sound_irq
    sta $0314
    lda #>peddle_sound_irq
    sta $0315
    lda #1
    sta peddle_sound_irq_installed
    cli
    rts

peddle_sound_shutdown:
    jsr peddle_sound_stop_all
    lda peddle_sound_irq_installed
    bne peddle_sound_shutdown_restore
    rts

peddle_sound_shutdown_restore:
    sei
    lda peddle_sound_old_irq_lo
    sta $0314
    lda peddle_sound_old_irq_hi
    sta $0315
    lda #0
    sta peddle_sound_irq_installed
    cli
    rts

peddle_sound_reset:
    jsr peddle_sound_stop_all
    lda #0
    sta peddle_sound_pool_used_lo
    sta peddle_sound_pool_used_hi
    sta peddle_sound_count

    lda peddle_sound_pool_header_lo
    ora peddle_sound_pool_header_hi
    beq peddle_sound_reset_slots

    lda peddle_sound_pool_header_lo
    sta ZP_PTR0_LO
    lda peddle_sound_pool_header_hi
    sta ZP_PTR0_HI
    ldy #2
    lda #0
    sta (ZP_PTR0_LO), y
    iny
    sta (ZP_PTR0_LO), y

peddle_sound_reset_slots:
    ldx #0
peddle_sound_reset_loop:
    lda #0
    sta peddle_sound_slot_inuse, x
    sta peddle_sound_slot_kind, x
    sta peddle_sound_slot_offset_lo, x
    sta peddle_sound_slot_offset_hi, x
    sta peddle_sound_slot_len_lo, x
    sta peddle_sound_slot_len_hi, x
    inx
    cpx #PEDDLE_SOUND_MAX
    bne peddle_sound_reset_loop
    rts

peddle_sound_stop_all:
    lda #0
    sta peddle_sound_playing
    sta peddle_sound_active_id
    sta peddle_sound_wait
    sta $d404
    sta $d40b
    sta $d412
    rts

peddle_sound_load:
    lda #0
    sta peddle_sound_load_return_id
    sta peddle_sound_load_return_id+1
    sta peddle_sound_load_return_err
    sta peddle_sound_load_return_err+1

    lda peddle_sound_initialized
    bne peddle_sound_load_is_initialized
    lda #1
    jmp peddle_sound_load_error_a

peddle_sound_load_is_initialized:
    lda peddle_sound_kind_hi
    bne peddle_sound_load_bad_kind
    lda peddle_sound_kind_lo
    cmp #PEDDLE_SOUND_REGSTREAM
    beq peddle_sound_load_kind_ok

peddle_sound_load_bad_kind:
    lda #2
    jmp peddle_sound_load_error_a

peddle_sound_load_kind_ok:
    jsr peddle_sound_prepare_data

    lda peddle_sound_data_len_lo
    ora peddle_sound_data_len_hi
    bne peddle_sound_load_nonempty
    lda #3
    jmp peddle_sound_load_error_a

peddle_sound_load_nonempty:
    jsr peddle_sound_validate_regstream
    bcc peddle_sound_load_valid
    lda #6
    jmp peddle_sound_load_error_a

peddle_sound_load_valid:
    jsr peddle_sound_find_free_slot
    bcc peddle_sound_load_has_slot
    lda #4
    jmp peddle_sound_load_error_a

peddle_sound_load_has_slot:
    jsr peddle_sound_check_pool_space
    bcc peddle_sound_load_has_space
    lda #5
    jmp peddle_sound_load_error_a

peddle_sound_load_has_space:
    jsr peddle_sound_copy_to_pool
    jsr peddle_sound_commit_slot

    lda peddle_sound_result_id
    sta peddle_sound_load_return_id
    lda #0
    sta peddle_sound_load_return_id+1
    sta peddle_sound_load_return_err
    sta peddle_sound_load_return_err+1
    rts

peddle_sound_load_error_a:
    sta peddle_sound_load_return_err
    lda #0
    sta peddle_sound_load_return_err+1
    sta peddle_sound_load_return_id
    sta peddle_sound_load_return_id+1
    rts

peddle_sound_prepare_data:
    lda peddle_sound_data_header_lo
    sta ZP_PTR0_LO
    lda peddle_sound_data_header_hi
    sta ZP_PTR0_HI

    ldy #2
    lda (ZP_PTR0_LO), y
    sta peddle_sound_data_len_lo
    iny
    lda (ZP_PTR0_LO), y
    sta peddle_sound_data_len_hi

    lda peddle_sound_data_header_lo
    clc
    adc #4
    sta peddle_sound_data_lo
    lda peddle_sound_data_header_hi
    adc #0
    sta peddle_sound_data_hi
    rts

peddle_sound_find_free_slot:
    ldx #0
peddle_sound_find_free_slot_loop:
    lda peddle_sound_slot_inuse, x
    beq peddle_sound_find_free_slot_found
    inx
    cpx #PEDDLE_SOUND_MAX
    bne peddle_sound_find_free_slot_loop
    sec
    rts

peddle_sound_find_free_slot_found:
    stx peddle_sound_slot_index
    txa
    clc
    adc #1
    sta peddle_sound_result_id
    clc
    rts

peddle_sound_check_pool_space:
    lda peddle_sound_pool_used_lo
    clc
    adc peddle_sound_data_len_lo
    sta peddle_sound_new_used_lo
    lda peddle_sound_pool_used_hi
    adc peddle_sound_data_len_hi
    sta peddle_sound_new_used_hi
    bcs peddle_sound_check_pool_full

    lda peddle_sound_pool_cap_hi
    cmp peddle_sound_new_used_hi
    bcc peddle_sound_check_pool_full
    bne peddle_sound_check_pool_ok
    lda peddle_sound_pool_cap_lo
    cmp peddle_sound_new_used_lo
    bcc peddle_sound_check_pool_full

peddle_sound_check_pool_ok:
    clc
    rts

peddle_sound_check_pool_full:
    sec
    rts

peddle_sound_copy_to_pool:
    lda peddle_sound_data_lo
    sta ZP_PTR0_LO
    lda peddle_sound_data_hi
    sta ZP_PTR0_HI

    lda peddle_sound_pool_data_lo
    clc
    adc peddle_sound_pool_used_lo
    sta ZP_PTR1_LO
    lda peddle_sound_pool_data_hi
    adc peddle_sound_pool_used_hi
    sta ZP_PTR1_HI

    lda peddle_sound_data_len_lo
    sta peddle_sound_copy_remaining_lo
    lda peddle_sound_data_len_hi
    sta peddle_sound_copy_remaining_hi

peddle_sound_copy_loop:
    lda peddle_sound_copy_remaining_lo
    ora peddle_sound_copy_remaining_hi
    beq peddle_sound_copy_done

    ldy #0
    lda (ZP_PTR0_LO), y
    sta (ZP_PTR1_LO), y

    inc ZP_PTR0_LO
    bne peddle_sound_copy_src_no_carry
    inc ZP_PTR0_HI
peddle_sound_copy_src_no_carry:
    inc ZP_PTR1_LO
    bne peddle_sound_copy_dst_no_carry
    inc ZP_PTR1_HI
peddle_sound_copy_dst_no_carry:
    lda peddle_sound_copy_remaining_lo
    bne peddle_sound_copy_dec_low
    dec peddle_sound_copy_remaining_hi
peddle_sound_copy_dec_low:
    dec peddle_sound_copy_remaining_lo
    jmp peddle_sound_copy_loop

peddle_sound_copy_done:
    rts

peddle_sound_commit_slot:
    ldx peddle_sound_slot_index
    lda #1
    sta peddle_sound_slot_inuse, x
    lda peddle_sound_kind_lo
    sta peddle_sound_slot_kind, x
    lda peddle_sound_pool_used_lo
    sta peddle_sound_slot_offset_lo, x
    lda peddle_sound_pool_used_hi
    sta peddle_sound_slot_offset_hi, x
    lda peddle_sound_data_len_lo
    sta peddle_sound_slot_len_lo, x
    lda peddle_sound_data_len_hi
    sta peddle_sound_slot_len_hi, x

    lda peddle_sound_new_used_lo
    sta peddle_sound_pool_used_lo
    lda peddle_sound_new_used_hi
    sta peddle_sound_pool_used_hi

    inc peddle_sound_count

    lda peddle_sound_pool_header_lo
    sta ZP_PTR0_LO
    lda peddle_sound_pool_header_hi
    sta ZP_PTR0_HI
    ldy #2
    lda peddle_sound_pool_used_lo
    sta (ZP_PTR0_LO), y
    iny
    lda peddle_sound_pool_used_hi
    sta (ZP_PTR0_LO), y
    rts

peddle_sound_validate_regstream:
    lda peddle_sound_data_lo
    sta ZP_PTR0_LO
    lda peddle_sound_data_hi
    sta ZP_PTR0_HI
    lda peddle_sound_data_len_lo
    sta peddle_sound_validate_remaining_lo
    lda peddle_sound_data_len_hi
    sta peddle_sound_validate_remaining_hi

peddle_sound_validate_loop:
    lda peddle_sound_validate_remaining_lo
    ora peddle_sound_validate_remaining_hi
    bne peddle_sound_validate_has_data
    sec
    rts

peddle_sound_validate_has_data:
    ldy #0
    lda (ZP_PTR0_LO), y
    beq peddle_sound_validate_ok
    cmp #1
    beq peddle_sound_validate_wait
    cmp #2
    beq peddle_sound_validate_write
    sec
    rts

peddle_sound_validate_wait:
    jsr peddle_sound_validate_need_2
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_2
    jmp peddle_sound_validate_loop

peddle_sound_validate_write:
    jsr peddle_sound_validate_need_3
    bcs peddle_sound_validate_bad
    ldy #1
    lda (ZP_PTR0_LO), y
    cmp #25
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_3
    jmp peddle_sound_validate_loop

peddle_sound_validate_ok:
    clc
    rts

peddle_sound_validate_bad:
    sec
    rts

peddle_sound_validate_need_2:
    lda peddle_sound_validate_remaining_hi
    bne peddle_sound_validate_need_ok
    lda peddle_sound_validate_remaining_lo
    cmp #2
    bcc peddle_sound_validate_need_bad
    clc
    rts

peddle_sound_validate_need_3:
    lda peddle_sound_validate_remaining_hi
    bne peddle_sound_validate_need_ok
    lda peddle_sound_validate_remaining_lo
    cmp #3
    bcc peddle_sound_validate_need_bad

peddle_sound_validate_need_ok:
    clc
    rts

peddle_sound_validate_need_bad:
    sec
    rts

peddle_sound_validate_advance_2:
    lda ZP_PTR0_LO
    clc
    adc #2
    sta ZP_PTR0_LO
    lda ZP_PTR0_HI
    adc #0
    sta ZP_PTR0_HI
    lda peddle_sound_validate_remaining_lo
    sec
    sbc #2
    sta peddle_sound_validate_remaining_lo
    lda peddle_sound_validate_remaining_hi
    sbc #0
    sta peddle_sound_validate_remaining_hi
    rts

peddle_sound_validate_advance_3:
    lda ZP_PTR0_LO
    clc
    adc #3
    sta ZP_PTR0_LO
    lda ZP_PTR0_HI
    adc #0
    sta ZP_PTR0_HI
    lda peddle_sound_validate_remaining_lo
    sec
    sbc #3
    sta peddle_sound_validate_remaining_lo
    lda peddle_sound_validate_remaining_hi
    sbc #0
    sta peddle_sound_validate_remaining_hi
    rts

peddle_sound_play:
    lda peddle_sound_handle_hi
    beq peddle_sound_play_handle_low
    rts

peddle_sound_play_handle_low:
    lda peddle_sound_handle_lo
    beq peddle_sound_play_done
    cmp #17
    bcc peddle_sound_play_range_ok
peddle_sound_play_done:
    rts

peddle_sound_play_range_ok:
    sec
    sbc #1
    tax
    lda peddle_sound_slot_inuse, x
    bne peddle_sound_play_slot_ok
    rts

peddle_sound_play_slot_ok:
    jsr peddle_sound_stop_all
    lda peddle_sound_handle_lo
    sta peddle_sound_active_id
    lda #0
    sta peddle_sound_wait

    lda peddle_sound_pool_data_lo
    clc
    adc peddle_sound_slot_offset_lo, x
    sta peddle_sound_pc_lo
    lda peddle_sound_pool_data_hi
    adc peddle_sound_slot_offset_hi, x
    sta peddle_sound_pc_hi

    lda #1
    sta peddle_sound_playing
    rts

peddle_sound_stop:
    lda peddle_sound_handle_hi
    beq peddle_sound_stop_handle_low
    rts

peddle_sound_stop_handle_low:
    lda peddle_sound_playing
    beq peddle_sound_stop_done
    lda peddle_sound_handle_lo
    cmp peddle_sound_active_id
    bne peddle_sound_stop_done
    jsr peddle_sound_stop_all
peddle_sound_stop_done:
    rts

peddle_sound_num:
    lda peddle_sound_count
    sta ZP_TMP0
    lda #0
    sta ZP_TMP1
    lda ZP_TMP0
    rts

peddle_sound_memfree:
    lda peddle_sound_pool_cap_lo
    sec
    sbc peddle_sound_pool_used_lo
    sta ZP_TMP0
    lda peddle_sound_pool_cap_hi
    sbc peddle_sound_pool_used_hi
    sta ZP_TMP1
    lda ZP_TMP0
    rts

peddle_sound_irq:
    pha
    txa
    pha
    tya
    pha

    lda ZP_PTR0_LO
    sta peddle_sound_irq_save_ptr0_lo
    lda ZP_PTR0_HI
    sta peddle_sound_irq_save_ptr0_hi

    lda peddle_sound_playing
    beq peddle_sound_irq_done

    lda peddle_sound_wait
    beq peddle_sound_irq_process_now
    dec peddle_sound_wait
    jmp peddle_sound_irq_done

peddle_sound_irq_process_now:
    jsr peddle_sound_irq_process

peddle_sound_irq_done:
    lda peddle_sound_irq_save_ptr0_lo
    sta ZP_PTR0_LO
    lda peddle_sound_irq_save_ptr0_hi
    sta ZP_PTR0_HI

    pla
    tay
    pla
    tax
    pla
    jmp (peddle_sound_old_irq_lo)

peddle_sound_irq_process:
    lda peddle_sound_pc_lo
    sta ZP_PTR0_LO
    lda peddle_sound_pc_hi
    sta ZP_PTR0_HI

peddle_sound_irq_command_loop:
    ldy #0
    lda (ZP_PTR0_LO), y
    beq peddle_sound_irq_end_stream
    cmp #1
    beq peddle_sound_irq_wait
    cmp #2
    beq peddle_sound_irq_write
    jmp peddle_sound_irq_end_stream

peddle_sound_irq_wait:
    ldy #1
    lda (ZP_PTR0_LO), y
    sta peddle_sound_wait
    jsr peddle_sound_irq_advance_2
    rts

peddle_sound_irq_write:
    ldy #1
    lda (ZP_PTR0_LO), y
    tax
    iny
    lda (ZP_PTR0_LO), y
    sta $d400,x
    jsr peddle_sound_irq_advance_3
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_end_stream:
    jsr peddle_sound_stop_all
    rts

peddle_sound_irq_advance_2:
    lda peddle_sound_pc_lo
    clc
    adc #2
    sta peddle_sound_pc_lo
    sta ZP_PTR0_LO
    lda peddle_sound_pc_hi
    adc #0
    sta peddle_sound_pc_hi
    sta ZP_PTR0_HI
    rts

peddle_sound_irq_advance_3:
    lda peddle_sound_pc_lo
    clc
    adc #3
    sta peddle_sound_pc_lo
    sta ZP_PTR0_LO
    lda peddle_sound_pc_hi
    adc #0
    sta peddle_sound_pc_hi
    sta ZP_PTR0_HI
    rts
`)
}
