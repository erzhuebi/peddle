package codegen

func (g *Generator) emitFileRuntime() {
	if !g.usedFileRuntime {
		return
	}

	g.emit(`
; file runtime
;
; V1 uses one normal logical file number for file streams and whole-buffer
; helpers. Use fileclose() before opening another stream.
;
; Public helpers emitted by codegen:
;   peddle_fileopen
;   peddle_fileclose
;   peddle_fileread
;   peddle_filewrite
;   peddle_fileload
;   peddle_filesave

KERNAL_READST = $ffb7
KERNAL_SETLFS = $ffba
KERNAL_SETNAM = $ffbd
KERNAL_OPEN   = $ffc0
KERNAL_CLOSE  = $ffc3
KERNAL_CHKIN  = $ffc6
KERNAL_CHKOUT = $ffc9
KERNAL_CLRCHN = $ffcc
KERNAL_CHRIN  = $ffcf
KERNAL_CHROUT = $ffd2

PEDDLE_FILE_LFN       = 1
PEDDLE_FILE_SECONDARY = 2
PEDDLE_FILE_NAME_MAX  = 120

peddle_file_name_lo:
    .byte 0
peddle_file_name_hi:
    .byte 0
peddle_file_name_len_lo:
    .byte 0
peddle_file_name_len_hi:
    .byte 0

peddle_file_mode_lo:
    .byte 0
peddle_file_mode_hi:
    .byte 0
peddle_file_mode_len_lo:
    .byte 0
peddle_file_mode_len_hi:
    .byte 0

peddle_file_buf_lo:
    .byte 0
peddle_file_buf_hi:
    .byte 0

peddle_file_max_lo:
    .byte 0
peddle_file_max_hi:
    .byte 0

peddle_file_limit_lo:
    .byte 0
peddle_file_limit_hi:
    .byte 0

peddle_file_count_lo:
    .byte 0
peddle_file_count_hi:
    .byte 0

peddle_file_device:
    .byte 8

peddle_file_handle:
    .byte 0

peddle_file_write_flag:
    .byte 0

peddle_file_error:
    .byte 0

peddle_file_last_char:
    .byte 0

peddle_file_name_len_eff:
    .byte 0

peddle_file_built_name_len:
    .byte 0

peddle_file_name_buffer:
    .fill 128, 0

peddle_fileopen:
    jsr peddle_file_mode_to_flag
    jsr peddle_file_open_current
    bcs peddle_fileopen_fail

    lda #PEDDLE_FILE_LFN
    sta peddle_file_handle
    rts

peddle_fileopen_fail:
    lda #0
    sta peddle_file_handle
    rts

peddle_fileload:
    lda #0
    sta peddle_file_write_flag

    jsr peddle_file_open_current
    bcc peddle_fileload_opened

    jsr peddle_file_return_minus_one
    rts

peddle_fileload_opened:
    lda #PEDDLE_FILE_LFN
    sta peddle_file_handle

    jsr peddle_fileread
    jsr peddle_file_close_lfn

    lda peddle_file_error
    beq peddle_fileload_success
    jsr peddle_file_return_minus_one
    rts

peddle_fileload_success:
    jsr peddle_file_return_count
    rts

peddle_filesave:
    lda #1
    sta peddle_file_write_flag

    jsr peddle_file_open_current
    bcc peddle_filesave_opened

    jsr peddle_file_return_minus_one
    rts

peddle_filesave_opened:
    lda #PEDDLE_FILE_LFN
    sta peddle_file_handle

    jsr peddle_filewrite
    jsr peddle_file_close_lfn

    lda peddle_file_error
    beq peddle_filesave_success
    jsr peddle_file_return_minus_one
    rts

peddle_filesave_success:
    jsr peddle_file_return_count
    rts

peddle_fileclose:
    jsr peddle_file_close_lfn
    rts

peddle_fileread:
    lda #0
    sta peddle_file_error

    jsr peddle_file_prepare_read_buffer

    lda peddle_file_handle
    beq peddle_fileread_error

    tax
    jsr KERNAL_CHKIN
    bcs peddle_fileread_error

peddle_fileread_loop:
    jsr peddle_file_count_reached_limit
    bne peddle_fileread_done

    jsr KERNAL_CHRIN
    sta peddle_file_last_char

    ldy #0
    sta (ZP_PTR1_LO), y

    jsr peddle_file_inc_data_ptr
    jsr peddle_file_inc_count

    jsr KERNAL_READST
    cmp #0
    beq peddle_fileread_loop

peddle_fileread_done:
    jsr KERNAL_CLRCHN
    jsr peddle_file_store_buffer_length
    jsr peddle_file_return_count
    rts

peddle_fileread_error:
    lda #1
    sta peddle_file_error
    jsr KERNAL_CLRCHN
    jsr peddle_file_store_buffer_length
    jsr peddle_file_return_minus_one
    rts

peddle_filewrite:
    lda #0
    sta peddle_file_error

    jsr peddle_file_prepare_write_buffer

    lda peddle_file_handle
    beq peddle_filewrite_error

    tax
    jsr KERNAL_CHKOUT
    bcs peddle_filewrite_error

peddle_filewrite_loop:
    jsr peddle_file_count_reached_limit
    bne peddle_filewrite_done

    ldy #0
    lda (ZP_PTR1_LO), y
    jsr KERNAL_CHROUT

    jsr peddle_file_inc_data_ptr
    jsr peddle_file_inc_count

    jsr KERNAL_READST
    cmp #0
    beq peddle_filewrite_loop

peddle_filewrite_done:
    jsr KERNAL_CLRCHN
    jsr peddle_file_return_count
    rts

peddle_filewrite_error:
    lda #1
    sta peddle_file_error
    jsr KERNAL_CLRCHN
    jsr peddle_file_return_minus_one
    rts

peddle_file_mode_to_flag:
    lda #0
    sta peddle_file_write_flag

    lda peddle_file_mode_len_lo
    ora peddle_file_mode_len_hi
    beq peddle_file_mode_done

    lda peddle_file_mode_lo
    sta ZP_PTR0_LO
    lda peddle_file_mode_hi
    sta ZP_PTR0_HI

    ldy #0
    lda (ZP_PTR0_LO), y
    cmp #87
    beq peddle_file_mode_write
    cmp #119
    beq peddle_file_mode_write
    rts

peddle_file_mode_write:
    lda #1
    sta peddle_file_write_flag

peddle_file_mode_done:
    rts

peddle_file_open_current:
    jsr peddle_file_build_name

    lda #PEDDLE_FILE_LFN
    jsr KERNAL_CLOSE
    jsr KERNAL_CLRCHN

    lda #PEDDLE_FILE_LFN
    ldx peddle_file_device
    ldy #PEDDLE_FILE_SECONDARY
    jsr KERNAL_SETLFS

    lda peddle_file_built_name_len
    ldx #<peddle_file_name_buffer
    ldy #>peddle_file_name_buffer
    jsr KERNAL_SETNAM

    jsr KERNAL_OPEN
    bcs peddle_file_open_fail

    jsr KERNAL_READST
    cmp #0
    bne peddle_file_open_fail

    clc
    rts

peddle_file_open_fail:
    lda #PEDDLE_FILE_LFN
    jsr KERNAL_CLOSE
    jsr KERNAL_CLRCHN
    sec
    rts

peddle_file_build_name:
    jsr peddle_file_effective_name_len

    lda peddle_file_name_lo
    sta ZP_PTR0_LO
    lda peddle_file_name_hi
    sta ZP_PTR0_HI

    lda peddle_file_write_flag
    bne peddle_file_build_write_name

    ldy #0

peddle_file_build_read_copy:
    cpy peddle_file_name_len_eff
    beq peddle_file_build_read_suffix

    lda (ZP_PTR0_LO), y
    sta peddle_file_name_buffer, y
    iny
    jmp peddle_file_build_read_copy

peddle_file_build_read_suffix:
    lda #44
    sta peddle_file_name_buffer, y
    iny
    lda #83
    sta peddle_file_name_buffer, y
    iny
    lda #44
    sta peddle_file_name_buffer, y
    iny
    lda #82
    sta peddle_file_name_buffer, y
    iny
    sty peddle_file_built_name_len
    rts

peddle_file_build_write_name:
    lda #64
    sta peddle_file_name_buffer
    lda #58
    sta peddle_file_name_buffer+1

    ldy #0

peddle_file_build_write_copy:
    cpy peddle_file_name_len_eff
    beq peddle_file_build_write_suffix

    lda (ZP_PTR0_LO), y
    sta peddle_file_name_buffer+2, y
    iny
    jmp peddle_file_build_write_copy

peddle_file_build_write_suffix:
    lda #44
    sta peddle_file_name_buffer+2, y
    iny
    lda #83
    sta peddle_file_name_buffer+2, y
    iny
    lda #44
    sta peddle_file_name_buffer+2, y
    iny
    lda #87
    sta peddle_file_name_buffer+2, y

    lda peddle_file_name_len_eff
    clc
    adc #6
    sta peddle_file_built_name_len
    rts

peddle_file_effective_name_len:
    lda peddle_file_name_len_hi
    bne peddle_file_name_len_cap

    lda peddle_file_name_len_lo
    cmp #PEDDLE_FILE_NAME_MAX+1
    bcc peddle_file_name_len_use_a

peddle_file_name_len_cap:
    lda #PEDDLE_FILE_NAME_MAX

peddle_file_name_len_use_a:
    sta peddle_file_name_len_eff
    rts

peddle_file_prepare_read_buffer:
    jsr peddle_file_prepare_buffer_common

    ; Clear destination array length before reading.
    lda peddle_file_buf_lo
    sta ZP_PTR0_LO
    lda peddle_file_buf_hi
    sta ZP_PTR0_HI

    ldy #2
    lda #0
    sta (ZP_PTR0_LO), y
    iny
    sta (ZP_PTR0_LO), y
    rts

peddle_file_prepare_write_buffer:
    jsr peddle_file_prepare_buffer_common
    rts

peddle_file_prepare_buffer_common:
    lda #0
    sta peddle_file_count_lo
    sta peddle_file_count_hi

    lda peddle_file_buf_lo
    sta ZP_PTR0_LO
    lda peddle_file_buf_hi
    sta ZP_PTR0_HI

    ldy #0
    lda (ZP_PTR0_LO), y
    sta peddle_file_limit_lo
    iny
    lda (ZP_PTR0_LO), y
    sta peddle_file_limit_hi

    lda peddle_file_max_hi
    cmp peddle_file_limit_hi
    bcc peddle_file_prepare_use_max
    bne peddle_file_prepare_limit_done

    lda peddle_file_max_lo
    cmp peddle_file_limit_lo
    bcc peddle_file_prepare_use_max
    jmp peddle_file_prepare_limit_done

peddle_file_prepare_use_max:
    lda peddle_file_max_lo
    sta peddle_file_limit_lo
    lda peddle_file_max_hi
    sta peddle_file_limit_hi

peddle_file_prepare_limit_done:
    lda peddle_file_buf_lo
    clc
    adc #4
    sta ZP_PTR1_LO
    lda peddle_file_buf_hi
    adc #0
    sta ZP_PTR1_HI
    rts

peddle_file_store_buffer_length:
    lda peddle_file_buf_lo
    sta ZP_PTR0_LO
    lda peddle_file_buf_hi
    sta ZP_PTR0_HI

    ldy #2
    lda peddle_file_count_lo
    sta (ZP_PTR0_LO), y
    iny
    lda peddle_file_count_hi
    sta (ZP_PTR0_LO), y
    rts

peddle_file_count_reached_limit:
    lda peddle_file_count_hi
    cmp peddle_file_limit_hi
    bcc peddle_file_count_not_reached
    bne peddle_file_count_reached

    lda peddle_file_count_lo
    cmp peddle_file_limit_lo
    bcc peddle_file_count_not_reached

peddle_file_count_reached:
    lda #1
    rts

peddle_file_count_not_reached:
    lda #0
    rts

peddle_file_inc_data_ptr:
    inc ZP_PTR1_LO
    bne peddle_file_inc_data_ptr_done
    inc ZP_PTR1_HI

peddle_file_inc_data_ptr_done:
    rts

peddle_file_inc_count:
    inc peddle_file_count_lo
    bne peddle_file_inc_count_done
    inc peddle_file_count_hi

peddle_file_inc_count_done:
    rts

peddle_file_close_lfn:
    lda peddle_file_handle
    beq peddle_file_close_done
    jsr KERNAL_CLOSE
    jsr KERNAL_CLRCHN
    lda #0
    sta peddle_file_handle

peddle_file_close_done:
    rts

peddle_file_return_count:
    lda peddle_file_count_lo
    sta ZP_TMP0
    lda peddle_file_count_hi
    sta ZP_TMP1
    lda ZP_TMP0
    rts

peddle_file_return_minus_one:
    lda #$ff
    sta ZP_TMP0
    sta ZP_TMP1
    lda ZP_TMP0
    rts
`)
}
