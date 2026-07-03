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
;   sound_play(id, voices, priority, flags) int
;   sound_stop(id)
;   sound_stop_voices(voices)
;   sound_num() int
;   sound_memfree() int
;
; Sound kind:
;   1 = unified timed SID stream
;
; Stream commands:
;   0                     end
;   1, frames             wait frames IRQ ticks
;   2, voice, note        set note frequency for logical voice
;   3, voice              gate off logical voice
;   4, voice, waveform    set waveform/control for logical voice
;   5, voice, ad          set attack/decay for logical voice
;   6, voice, sr          set sustain/release for logical voice
;   7, volume             set global volume
;   8, voice, lo, hi      set raw frequency for logical voice
;   9, reg, value         raw write to $D400 + reg

PEDDLE_SOUND_STREAM     = 1
PEDDLE_SOUND_MAX        = 16
PEDDLE_SOUND_PLAYERS    = 4
PEDDLE_SOUND_NOTE_COUNT = 85
PEDDLE_SOUND_VOICE1     = 1
PEDDLE_SOUND_VOICE2     = 2
PEDDLE_SOUND_VOICE3     = 4
PEDDLE_SOUND_ALL        = 7
PEDDLE_SOUND_REPLACE    = 1
PEDDLE_SOUND_OVERLAY    = 2
PEDDLE_SOUND_LOOP       = 4

peddle_sound_load_return_id:
    .fill 2, 0
peddle_sound_load_return_err:
    .fill 2, 0
peddle_sound_play_return_err:
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
peddle_sound_play_voices_lo:
    .byte 0
peddle_sound_play_voices_hi:
    .byte 0
peddle_sound_play_priority_lo:
    .byte 0
peddle_sound_play_priority_hi:
    .byte 0
peddle_sound_play_flags_lo:
    .byte 0
peddle_sound_play_flags_hi:
    .byte 0
peddle_sound_stop_voices_lo:
    .byte 0
peddle_sound_stop_voices_hi:
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
peddle_sound_player_index:
    .byte 0
peddle_sound_irq_player_index:
    .byte 0
peddle_sound_stop_index:
    .byte 0
peddle_sound_stop_owner:
    .byte 0
peddle_sound_command_voice:
    .byte 0
peddle_sound_raw_reg:
    .byte 0
peddle_sound_voice_base:
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

peddle_sound_player_inuse:
    .fill 4, 0
peddle_sound_player_id:
    .fill 4, 0
peddle_sound_player_wait:
    .fill 4, 0
peddle_sound_player_pc_lo:
    .fill 4, 0
peddle_sound_player_pc_hi:
    .fill 4, 0
peddle_sound_player_voices:
    .fill 4, 0
peddle_sound_player_priority:
    .fill 4, 0
peddle_sound_player_flags:
    .fill 4, 0
peddle_sound_voice_owner:
    .fill 3, 0

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
    sta peddle_sound_voice_base
    sta peddle_sound_voice_owner
    sta peddle_sound_voice_owner+1
    sta peddle_sound_voice_owner+2
    ldx #0
peddle_sound_stop_all_loop:
    sta peddle_sound_player_inuse, x
    sta peddle_sound_player_id, x
    sta peddle_sound_player_wait, x
    sta peddle_sound_player_pc_lo, x
    sta peddle_sound_player_pc_hi, x
    sta peddle_sound_player_voices, x
    sta peddle_sound_player_priority, x
    sta peddle_sound_player_flags, x
    inx
    cpx #PEDDLE_SOUND_PLAYERS
    bne peddle_sound_stop_all_loop
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
    cmp #PEDDLE_SOUND_STREAM
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
    jsr peddle_sound_validate_stream
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

peddle_sound_validate_stream:
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
    beq peddle_sound_validate_end
    cmp #1
    beq peddle_sound_validate_wait
    cmp #2
    beq peddle_sound_validate_note
    cmp #3
    beq peddle_sound_validate_gate_off
    cmp #4
    beq peddle_sound_validate_waveform
    cmp #5
    beq peddle_sound_validate_ad
    cmp #6
    beq peddle_sound_validate_sr
    cmp #7
    beq peddle_sound_validate_volume
    cmp #8
    beq peddle_sound_validate_freq
    cmp #9
    beq peddle_sound_validate_raw
    sec
    rts

peddle_sound_validate_end:
    clc
    rts

peddle_sound_validate_wait:
    jsr peddle_sound_validate_need_2
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_2
    jmp peddle_sound_validate_loop

peddle_sound_validate_note:
    jsr peddle_sound_validate_need_3
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_voice
    bcs peddle_sound_validate_bad
    ldy #2
    lda (ZP_PTR0_LO), y
    cmp #PEDDLE_SOUND_NOTE_COUNT
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_3
    jmp peddle_sound_validate_loop

peddle_sound_validate_gate_off:
    jsr peddle_sound_validate_need_2
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_voice
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_2
    jmp peddle_sound_validate_loop

peddle_sound_validate_waveform:
peddle_sound_validate_ad:
peddle_sound_validate_sr:
    jsr peddle_sound_validate_need_3
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_voice
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_3
    jmp peddle_sound_validate_loop

peddle_sound_validate_volume:
    jsr peddle_sound_validate_need_2
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_2
    jmp peddle_sound_validate_loop

peddle_sound_validate_freq:
    jsr peddle_sound_validate_need_4
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_voice
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_4
    jmp peddle_sound_validate_loop

peddle_sound_validate_raw:
    jsr peddle_sound_validate_need_3
    bcs peddle_sound_validate_bad
    ldy #1
    lda (ZP_PTR0_LO), y
    cmp #25
    bcs peddle_sound_validate_bad
    jsr peddle_sound_validate_advance_3
    jmp peddle_sound_validate_loop

peddle_sound_validate_voice:
    ldy #1
    lda (ZP_PTR0_LO), y
    cmp #3
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
    clc
    rts

peddle_sound_validate_need_4:
    lda peddle_sound_validate_remaining_hi
    bne peddle_sound_validate_need_ok
    lda peddle_sound_validate_remaining_lo
    cmp #4
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

peddle_sound_validate_advance_4:
    lda ZP_PTR0_LO
    clc
    adc #4
    sta ZP_PTR0_LO
    lda ZP_PTR0_HI
    adc #0
    sta ZP_PTR0_HI
    lda peddle_sound_validate_remaining_lo
    sec
    sbc #4
    sta peddle_sound_validate_remaining_lo
    lda peddle_sound_validate_remaining_hi
    sbc #0
    sta peddle_sound_validate_remaining_hi
    rts

peddle_sound_play:
    lda #0
    sta peddle_sound_play_return_err
    sta peddle_sound_play_return_err+1

    lda peddle_sound_initialized
    bne peddle_sound_play_initialized
    lda #1
    jmp peddle_sound_play_error_a

peddle_sound_play_initialized:
    lda peddle_sound_handle_hi
    beq peddle_sound_play_handle_low
    lda #2
    jmp peddle_sound_play_error_a

peddle_sound_play_handle_low:
    lda peddle_sound_handle_lo
    beq peddle_sound_play_bad_handle
    cmp #17
    bcc peddle_sound_play_range_ok
peddle_sound_play_bad_handle:
    lda #2
    jmp peddle_sound_play_error_a

peddle_sound_play_range_ok:
    sec
    sbc #1
    tax
    lda peddle_sound_slot_inuse, x
    bne peddle_sound_play_slot_ok
    lda #2
    jmp peddle_sound_play_error_a

peddle_sound_play_slot_ok:
    lda peddle_sound_play_voices_hi
    beq peddle_sound_play_voices_low
    lda #3
    jmp peddle_sound_play_error_a
peddle_sound_play_voices_low:
    lda peddle_sound_play_voices_lo
    beq peddle_sound_play_bad_voices
    and #248
    beq peddle_sound_play_flags_check
peddle_sound_play_bad_voices:
    lda #3
    jmp peddle_sound_play_error_a

peddle_sound_play_flags_check:
    lda peddle_sound_play_flags_hi
    beq peddle_sound_play_flags_low
    lda #6
    jmp peddle_sound_play_error_a
peddle_sound_play_flags_low:
    lda peddle_sound_play_flags_lo
    and #248
    beq peddle_sound_play_mode
    lda #6
    jmp peddle_sound_play_error_a

peddle_sound_play_mode:
    lda peddle_sound_play_flags_lo
    and #PEDDLE_SOUND_REPLACE
    beq peddle_sound_play_not_replace
    jsr peddle_sound_stop_all
    jmp peddle_sound_play_allocate

peddle_sound_play_not_replace:
    jsr peddle_sound_play_check_conflicts
    bcc peddle_sound_play_conflicts_ok
    lda #5
    jmp peddle_sound_play_error_a
peddle_sound_play_conflicts_ok:
    lda peddle_sound_play_flags_lo
    and #PEDDLE_SOUND_OVERLAY
    beq peddle_sound_play_allocate
    jsr peddle_sound_play_stop_conflicts

peddle_sound_play_allocate:
    ldx #0
peddle_sound_play_find_player_loop:
    lda peddle_sound_player_inuse, x
    beq peddle_sound_play_found_player
    inx
    cpx #PEDDLE_SOUND_PLAYERS
    bne peddle_sound_play_find_player_loop
    lda #4
    jmp peddle_sound_play_error_a

peddle_sound_play_found_player:
    stx peddle_sound_player_index
    lda peddle_sound_handle_lo
    sec
    sbc #1
    tax

    lda peddle_sound_pool_data_lo
    clc
    adc peddle_sound_slot_offset_lo, x
    sta peddle_sound_data_lo
    lda peddle_sound_pool_data_hi
    adc peddle_sound_slot_offset_hi, x
    sta peddle_sound_data_hi

    ldx peddle_sound_player_index
    lda #1
    sta peddle_sound_player_inuse, x
    lda peddle_sound_handle_lo
    sta peddle_sound_player_id, x
    lda #0
    sta peddle_sound_player_wait, x
    lda peddle_sound_data_lo
    sta peddle_sound_player_pc_lo, x
    lda peddle_sound_data_hi
    sta peddle_sound_player_pc_hi, x
    lda peddle_sound_play_voices_lo
    sta peddle_sound_player_voices, x
    lda peddle_sound_play_priority_lo
    sta peddle_sound_player_priority, x
    lda peddle_sound_play_flags_lo
    sta peddle_sound_player_flags, x

    txa
    clc
    adc #1
    sta peddle_sound_stop_owner

    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE1
    beq peddle_sound_play_own_voice2
    lda peddle_sound_stop_owner
    sta peddle_sound_voice_owner
peddle_sound_play_own_voice2:
    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE2
    beq peddle_sound_play_own_voice3
    lda peddle_sound_stop_owner
    sta peddle_sound_voice_owner+1
peddle_sound_play_own_voice3:
    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE3
    beq peddle_sound_play_started
    lda peddle_sound_stop_owner
    sta peddle_sound_voice_owner+2

peddle_sound_play_started:
    lda #1
    sta peddle_sound_playing
    lda #0
    sta peddle_sound_play_return_err
    sta peddle_sound_play_return_err+1
    sta ZP_TMP0
    sta ZP_TMP1
    rts

peddle_sound_play_error_a:
    sta peddle_sound_play_return_err
    sta ZP_TMP0
    lda #0
    sta peddle_sound_play_return_err+1
    sta ZP_TMP1
    lda ZP_TMP0
    rts

peddle_sound_play_check_conflicts:
    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE1
    beq peddle_sound_play_check_voice2
    ldx peddle_sound_voice_owner
    beq peddle_sound_play_check_voice2
    jsr peddle_sound_play_can_take_owner_x
    bcs peddle_sound_play_conflict_busy
peddle_sound_play_check_voice2:
    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE2
    beq peddle_sound_play_check_voice3
    ldx peddle_sound_voice_owner+1
    beq peddle_sound_play_check_voice3
    jsr peddle_sound_play_can_take_owner_x
    bcs peddle_sound_play_conflict_busy
peddle_sound_play_check_voice3:
    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE3
    beq peddle_sound_play_no_conflict
    ldx peddle_sound_voice_owner+2
    beq peddle_sound_play_no_conflict
    jsr peddle_sound_play_can_take_owner_x
    bcs peddle_sound_play_conflict_busy
peddle_sound_play_no_conflict:
    clc
    rts
peddle_sound_play_conflict_busy:
    sec
    rts

peddle_sound_play_can_take_owner_x:
    lda peddle_sound_play_flags_lo
    and #PEDDLE_SOUND_OVERLAY
    beq peddle_sound_play_owner_busy
    txa
    sec
    sbc #1
    tax
    lda peddle_sound_play_priority_lo
    cmp peddle_sound_player_priority, x
    bcc peddle_sound_play_owner_busy
    clc
    rts
peddle_sound_play_owner_busy:
    sec
    rts

peddle_sound_play_stop_conflicts:
    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE1
    beq peddle_sound_play_stop_voice2
    ldx peddle_sound_voice_owner
    beq peddle_sound_play_stop_voice2
    dex
    lda peddle_sound_player_voices, x
    and #254
    sta peddle_sound_player_voices, x
    lda #0
    sta peddle_sound_voice_owner
    sta $d404
    lda peddle_sound_player_voices, x
    bne peddle_sound_play_stop_voice2
    jsr peddle_sound_stop_player_x
peddle_sound_play_stop_voice2:
    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE2
    beq peddle_sound_play_stop_voice3
    ldx peddle_sound_voice_owner+1
    beq peddle_sound_play_stop_voice3
    dex
    lda peddle_sound_player_voices, x
    and #253
    sta peddle_sound_player_voices, x
    lda #0
    sta peddle_sound_voice_owner+1
    sta $d40b
    lda peddle_sound_player_voices, x
    bne peddle_sound_play_stop_voice3
    jsr peddle_sound_stop_player_x
peddle_sound_play_stop_voice3:
    lda peddle_sound_play_voices_lo
    and #PEDDLE_SOUND_VOICE3
    beq peddle_sound_play_stop_done
    ldx peddle_sound_voice_owner+2
    beq peddle_sound_play_stop_done
    dex
    lda peddle_sound_player_voices, x
    and #251
    sta peddle_sound_player_voices, x
    lda #0
    sta peddle_sound_voice_owner+2
    sta $d412
    lda peddle_sound_player_voices, x
    bne peddle_sound_play_stop_done
    jsr peddle_sound_stop_player_x
peddle_sound_play_stop_done:
    rts

peddle_sound_stop:
    lda peddle_sound_handle_hi
    beq peddle_sound_stop_handle_low
    rts

peddle_sound_stop_handle_low:
    lda peddle_sound_handle_lo
    beq peddle_sound_stop_done
    ldx #0
peddle_sound_stop_loop:
    stx peddle_sound_stop_index
    lda peddle_sound_player_inuse, x
    beq peddle_sound_stop_next
    lda peddle_sound_player_id, x
    cmp peddle_sound_handle_lo
    bne peddle_sound_stop_next
    jsr peddle_sound_stop_player_x
peddle_sound_stop_next:
    ldx peddle_sound_stop_index
    inx
    cpx #PEDDLE_SOUND_PLAYERS
    bne peddle_sound_stop_loop
peddle_sound_stop_done:
    rts

peddle_sound_stop_voices:
    lda peddle_sound_stop_voices_hi
    beq peddle_sound_stop_voices_low
    rts

peddle_sound_stop_voices_low:
    lda peddle_sound_stop_voices_lo
    beq peddle_sound_stop_voices_done
    and #248
    beq peddle_sound_stop_voices_apply
    rts

peddle_sound_stop_voices_apply:
    lda peddle_sound_stop_voices_lo
    and #PEDDLE_SOUND_VOICE1
    beq peddle_sound_stop_voices_voice2
    ldx peddle_sound_voice_owner
    beq peddle_sound_stop_voices_voice2
    dex
    lda peddle_sound_player_voices, x
    and #254
    sta peddle_sound_player_voices, x
    lda #0
    sta peddle_sound_voice_owner
    sta $d404
    lda peddle_sound_player_voices, x
    bne peddle_sound_stop_voices_voice2
    jsr peddle_sound_stop_player_x
peddle_sound_stop_voices_voice2:
    lda peddle_sound_stop_voices_lo
    and #PEDDLE_SOUND_VOICE2
    beq peddle_sound_stop_voices_voice3
    ldx peddle_sound_voice_owner+1
    beq peddle_sound_stop_voices_voice3
    dex
    lda peddle_sound_player_voices, x
    and #253
    sta peddle_sound_player_voices, x
    lda #0
    sta peddle_sound_voice_owner+1
    sta $d40b
    lda peddle_sound_player_voices, x
    bne peddle_sound_stop_voices_voice3
    jsr peddle_sound_stop_player_x
peddle_sound_stop_voices_voice3:
    lda peddle_sound_stop_voices_lo
    and #PEDDLE_SOUND_VOICE3
    beq peddle_sound_stop_voices_done
    ldx peddle_sound_voice_owner+2
    beq peddle_sound_stop_voices_done
    dex
    lda peddle_sound_player_voices, x
    and #251
    sta peddle_sound_player_voices, x
    lda #0
    sta peddle_sound_voice_owner+2
    sta $d412
    lda peddle_sound_player_voices, x
    bne peddle_sound_stop_voices_done
    jsr peddle_sound_stop_player_x
peddle_sound_stop_voices_done:
    rts

peddle_sound_stop_player_x:
    stx peddle_sound_stop_index
    txa
    clc
    adc #1
    sta peddle_sound_stop_owner

    lda #0
    sta peddle_sound_player_inuse, x
    sta peddle_sound_player_id, x
    sta peddle_sound_player_wait, x
    sta peddle_sound_player_pc_lo, x
    sta peddle_sound_player_pc_hi, x
    sta peddle_sound_player_voices, x
    sta peddle_sound_player_priority, x
    sta peddle_sound_player_flags, x

    lda peddle_sound_voice_owner
    cmp peddle_sound_stop_owner
    bne peddle_sound_stop_player_voice2
    lda #0
    sta peddle_sound_voice_owner
    sta $d404
peddle_sound_stop_player_voice2:
    lda peddle_sound_voice_owner+1
    cmp peddle_sound_stop_owner
    bne peddle_sound_stop_player_voice3
    lda #0
    sta peddle_sound_voice_owner+1
    sta $d40b
peddle_sound_stop_player_voice3:
    lda peddle_sound_voice_owner+2
    cmp peddle_sound_stop_owner
    bne peddle_sound_stop_player_refresh
    lda #0
    sta peddle_sound_voice_owner+2
    sta $d412
peddle_sound_stop_player_refresh:
    jsr peddle_sound_refresh_playing
    ldx peddle_sound_stop_index
    rts

peddle_sound_refresh_playing:
    lda #0
    sta peddle_sound_playing
    ldx #0
peddle_sound_refresh_playing_loop:
    lda peddle_sound_player_inuse, x
    beq peddle_sound_refresh_playing_next
    lda #1
    sta peddle_sound_playing
    rts
peddle_sound_refresh_playing_next:
    inx
    cpx #PEDDLE_SOUND_PLAYERS
    bne peddle_sound_refresh_playing_loop
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

    ldx #0
peddle_sound_irq_player_loop:
    stx peddle_sound_irq_player_index
    lda peddle_sound_player_inuse, x
    beq peddle_sound_irq_next_player
    lda peddle_sound_player_wait, x
    beq peddle_sound_irq_process_now
    dec peddle_sound_player_wait, x
    jmp peddle_sound_irq_next_player

peddle_sound_irq_process_now:
    jsr peddle_sound_irq_process
peddle_sound_irq_next_player:
    ldx peddle_sound_irq_player_index
    inx
    cpx #PEDDLE_SOUND_PLAYERS
    bne peddle_sound_irq_player_loop

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
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_pc_lo, x
    sta ZP_PTR0_LO
    lda peddle_sound_player_pc_hi, x
    sta ZP_PTR0_HI

peddle_sound_irq_command_loop:
    ldy #0
    lda (ZP_PTR0_LO), y
    bne peddle_sound_irq_not_end
    jmp peddle_sound_irq_end_stream
peddle_sound_irq_not_end:
    cmp #1
    bne peddle_sound_irq_not_wait
    jmp peddle_sound_irq_wait
peddle_sound_irq_not_wait:
    cmp #2
    bne peddle_sound_irq_not_note
    jmp peddle_sound_irq_note
peddle_sound_irq_not_note:
    cmp #3
    bne peddle_sound_irq_not_gate_off
    jmp peddle_sound_irq_gate_off
peddle_sound_irq_not_gate_off:
    cmp #4
    bne peddle_sound_irq_not_waveform
    jmp peddle_sound_irq_waveform
peddle_sound_irq_not_waveform:
    cmp #5
    bne peddle_sound_irq_not_ad
    jmp peddle_sound_irq_ad
peddle_sound_irq_not_ad:
    cmp #6
    bne peddle_sound_irq_not_sr
    jmp peddle_sound_irq_sr
peddle_sound_irq_not_sr:
    cmp #7
    bne peddle_sound_irq_not_volume
    jmp peddle_sound_irq_volume
peddle_sound_irq_not_volume:
    cmp #8
    bne peddle_sound_irq_not_freq
    jmp peddle_sound_irq_freq
peddle_sound_irq_not_freq:
    cmp #9
    bne peddle_sound_irq_unknown
    jmp peddle_sound_irq_raw
peddle_sound_irq_unknown:
    jmp peddle_sound_irq_end_stream

peddle_sound_irq_wait:
    ldy #1
    lda (ZP_PTR0_LO), y
    ldx peddle_sound_irq_player_index
    sta peddle_sound_player_wait, x
    jsr peddle_sound_irq_advance_2
    rts

peddle_sound_irq_note:
    ldy #1
    lda (ZP_PTR0_LO), y
    jsr peddle_sound_irq_voice_allowed
    beq peddle_sound_irq_note_skip
    lda peddle_sound_command_voice
    jsr peddle_sound_voice_to_base
    sta peddle_sound_voice_base
    ldy #2
    lda (ZP_PTR0_LO), y
    tax
    lda peddle_sound_note_lo, x
    ldy peddle_sound_voice_base
    sta $d400, y
    lda peddle_sound_note_hi, x
    ldy peddle_sound_voice_base
    iny
    sta $d400, y
peddle_sound_irq_note_skip:
    jsr peddle_sound_irq_advance_3
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_gate_off:
    ldy #1
    lda (ZP_PTR0_LO), y
    jsr peddle_sound_irq_voice_allowed
    beq peddle_sound_irq_gate_off_skip
    lda peddle_sound_command_voice
    jsr peddle_sound_voice_to_base
    clc
    adc #4
    tax
    lda #0
    sta $d400, x
peddle_sound_irq_gate_off_skip:
    jsr peddle_sound_irq_advance_2
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_waveform:
    ldy #1
    lda (ZP_PTR0_LO), y
    jsr peddle_sound_irq_voice_allowed
    beq peddle_sound_irq_waveform_skip
    lda peddle_sound_command_voice
    jsr peddle_sound_voice_to_base
    clc
    adc #4
    tax
    ldy #2
    lda (ZP_PTR0_LO), y
    sta $d400, x
peddle_sound_irq_waveform_skip:
    jsr peddle_sound_irq_advance_3
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_ad:
    ldy #1
    lda (ZP_PTR0_LO), y
    jsr peddle_sound_irq_voice_allowed
    beq peddle_sound_irq_ad_skip
    lda peddle_sound_command_voice
    jsr peddle_sound_voice_to_base
    clc
    adc #5
    tax
    ldy #2
    lda (ZP_PTR0_LO), y
    sta $d400, x
peddle_sound_irq_ad_skip:
    jsr peddle_sound_irq_advance_3
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_sr:
    ldy #1
    lda (ZP_PTR0_LO), y
    jsr peddle_sound_irq_voice_allowed
    beq peddle_sound_irq_sr_skip
    lda peddle_sound_command_voice
    jsr peddle_sound_voice_to_base
    clc
    adc #6
    tax
    ldy #2
    lda (ZP_PTR0_LO), y
    sta $d400, x
peddle_sound_irq_sr_skip:
    jsr peddle_sound_irq_advance_3
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_volume:
    ldy #1
    lda (ZP_PTR0_LO), y
    sta $d418
    jsr peddle_sound_irq_advance_2
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_freq:
    ldy #1
    lda (ZP_PTR0_LO), y
    jsr peddle_sound_irq_voice_allowed
    beq peddle_sound_irq_freq_skip
    lda peddle_sound_command_voice
    jsr peddle_sound_voice_to_base
    tax
    ldy #2
    lda (ZP_PTR0_LO), y
    sta $d400, x
    inx
    ldy #3
    lda (ZP_PTR0_LO), y
    sta $d400, x
peddle_sound_irq_freq_skip:
    jsr peddle_sound_irq_advance_4
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_raw:
    ldy #1
    lda (ZP_PTR0_LO), y
    jsr peddle_sound_irq_raw_allowed
    beq peddle_sound_irq_raw_skip
    lda peddle_sound_raw_reg
    tax
    ldy #2
    lda (ZP_PTR0_LO), y
    sta $d400, x
peddle_sound_irq_raw_skip:
    jsr peddle_sound_irq_advance_3
    jmp peddle_sound_irq_command_loop

peddle_sound_irq_end_stream:
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_flags, x
    and #PEDDLE_SOUND_LOOP
    beq peddle_sound_irq_stop_stream
    jsr peddle_sound_restart_player_x
    rts
peddle_sound_irq_stop_stream:
    jsr peddle_sound_stop_player_x
    rts

peddle_sound_irq_voice_allowed:
    sta peddle_sound_command_voice
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_voices, x
    ldx peddle_sound_command_voice
    cpx #1
    beq peddle_sound_irq_voice_allowed_1
    cpx #2
    beq peddle_sound_irq_voice_allowed_2
    and #PEDDLE_SOUND_VOICE1
    rts
peddle_sound_irq_voice_allowed_1:
    and #PEDDLE_SOUND_VOICE2
    rts
peddle_sound_irq_voice_allowed_2:
    and #PEDDLE_SOUND_VOICE3
    rts

peddle_sound_irq_raw_allowed:
    sta peddle_sound_raw_reg
    cmp #7
    bcc peddle_sound_irq_raw_voice0
    cmp #14
    bcc peddle_sound_irq_raw_voice1
    cmp #21
    bcc peddle_sound_irq_raw_voice2
    lda #1
    rts
peddle_sound_irq_raw_voice0:
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_voices, x
    and #PEDDLE_SOUND_VOICE1
    rts
peddle_sound_irq_raw_voice1:
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_voices, x
    and #PEDDLE_SOUND_VOICE2
    rts
peddle_sound_irq_raw_voice2:
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_voices, x
    and #PEDDLE_SOUND_VOICE3
    rts

peddle_sound_restart_player_x:
    stx peddle_sound_player_index
    lda peddle_sound_player_id, x
    sec
    sbc #1
    tax
    lda peddle_sound_pool_data_lo
    clc
    adc peddle_sound_slot_offset_lo, x
    sta peddle_sound_data_lo
    lda peddle_sound_pool_data_hi
    adc peddle_sound_slot_offset_hi, x
    sta peddle_sound_data_hi
    ldx peddle_sound_player_index
    lda peddle_sound_data_lo
    sta peddle_sound_player_pc_lo, x
    lda peddle_sound_data_hi
    sta peddle_sound_player_pc_hi, x
    lda #0
    sta peddle_sound_player_wait, x
    rts

peddle_sound_voice_to_base:
    cmp #1
    beq peddle_sound_voice_base_1
    cmp #2
    beq peddle_sound_voice_base_2
    lda #0
    rts
peddle_sound_voice_base_1:
    lda #7
    rts
peddle_sound_voice_base_2:
    lda #14
    rts

peddle_sound_irq_advance_2:
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_pc_lo, x
    clc
    adc #2
    sta peddle_sound_player_pc_lo, x
    sta ZP_PTR0_LO
    lda peddle_sound_player_pc_hi, x
    adc #0
    sta peddle_sound_player_pc_hi, x
    sta ZP_PTR0_HI
    rts

peddle_sound_irq_advance_3:
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_pc_lo, x
    clc
    adc #3
    sta peddle_sound_player_pc_lo, x
    sta ZP_PTR0_LO
    lda peddle_sound_player_pc_hi, x
    adc #0
    sta peddle_sound_player_pc_hi, x
    sta ZP_PTR0_HI
    rts

peddle_sound_irq_advance_4:
    ldx peddle_sound_irq_player_index
    lda peddle_sound_player_pc_lo, x
    clc
    adc #4
    sta peddle_sound_player_pc_lo, x
    sta ZP_PTR0_LO
    lda peddle_sound_player_pc_hi, x
    adc #0
    sta peddle_sound_player_pc_hi, x
    sta ZP_PTR0_HI
    rts

peddle_sound_note_lo:
    .byte 22,39,57,75,95,116,138,161,186,212,240,14,45,78,113,150
    .byte 190,231,20,66,116,169,224,27,90,156,226,45,123,207,39,133
    .byte 232,81,193,55,180,56,196,89,247,157,78,10,208,162,129,109
    .byte 103,112,137,178,237,59,156,19,160,69,2,218,206,224,17,100
    .byte 218,118,57,38,64,137,4,180,156,192,35,200,180,235,114,76
    .byte 128,18,8,104,57
peddle_sound_note_hi:
    .byte 1,1,1,1,1,1,1,1,1,1,1,2,2,2,2,2
    .byte 2,2,3,3,3,3,3,4,4,4,4,5,5,5,6,6
    .byte 6,7,7,8,8,9,9,10,10,11,12,13,13,14,15,16
    .byte 17,18,19,20,21,23,24,26,27,29,31,32,34,36,39,41
    .byte 43,46,49,52,55,58,62,65,69,73,78,82,87,92,98,104
    .byte 110,117,124,131,139
`)
}
